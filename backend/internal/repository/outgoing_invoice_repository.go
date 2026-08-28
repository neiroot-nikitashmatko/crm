package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrOutgoingInvoiceNotFound  = errors.New("расходная накладная не найдена")
	ErrOutgoingInvoiceDealTaken = errors.New("по этой сделке уже есть расходная накладная")
)

type OutgoingInvoiceRepository struct {
	db *pgxpool.Pool
}

func NewOutgoingInvoiceRepository(db *pgxpool.Pool) *OutgoingInvoiceRepository {
	return &OutgoingInvoiceRepository{db: db}
}

const outgoingInvoiceSelect = `
SELECT
  id::text,
  invoice_number,
  invoice_date,
  deal_id::text,
  total,
  comment,
  created_at
FROM outgoing_invoices
`

func (r *OutgoingInvoiceRepository) List(ctx context.Context) ([]model.OutgoingInvoice, error) {
	query := outgoingInvoiceSelect + `
ORDER BY invoice_date DESC, created_at DESC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]model.OutgoingInvoice, 0)
	for rows.Next() {
		item, err := scanOutgoingInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	itemsByInvoice, err := listInvoiceItemsByKind(ctx, r.db, "outgoing")
	if err != nil {
		return nil, err
	}
	for index := range invoices {
		items := itemsByInvoice[invoices[index].ID]
		if items == nil {
			items = make([]model.InvoiceItem, 0)
		}
		invoices[index].Items = items
	}
	return invoices, nil
}

func (r *OutgoingInvoiceRepository) ExistsByDealID(ctx context.Context, dealID string, excludeInvoiceID string) (bool, error) {
	const query = `
SELECT EXISTS (
  SELECT 1
  FROM outgoing_invoices
  WHERE deal_id = $1::uuid
    AND ($2::uuid IS NULL OR id <> $2::uuid)
)
`
	var exclude any
	if strings.TrimSpace(excludeInvoiceID) != "" {
		exclude = excludeInvoiceID
	}
	var exists bool
	if err := r.db.QueryRow(ctx, query, dealID, exclude).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *OutgoingInvoiceRepository) Create(ctx context.Context, input model.UpsertOutgoingInvoiceInput, total float64) (model.OutgoingInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.OutgoingInvoice{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
INSERT INTO outgoing_invoices (invoice_date, deal_id, total, comment)
VALUES (to_timestamp($1::double precision / 1000.0), $2::uuid, $3, $4)
RETURNING
  id::text,
  invoice_number,
  invoice_date,
  deal_id::text,
  total,
  comment,
  created_at
`
	invoice, err := scanOutgoingInvoice(tx.QueryRow(ctx, query, input.Date, input.DealID, total, input.Comment))
	if err != nil {
		if isUniqueViolation(err) {
			return model.OutgoingInvoice{}, ErrOutgoingInvoiceDealTaken
		}
		if isForeignKeyViolation(err) {
			return model.OutgoingInvoice{}, errors.New("сделка не найдена")
		}
		return model.OutgoingInvoice{}, err
	}

	if err := replaceInvoiceItems(ctx, tx, "outgoing", invoice.ID, input.Items); err != nil {
		return model.OutgoingInvoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.OutgoingInvoice{}, err
	}

	invoice.Items = append([]model.InvoiceItem(nil), input.Items...)
	return invoice, nil
}

func (r *OutgoingInvoiceRepository) Update(ctx context.Context, id string, input model.UpsertOutgoingInvoiceInput, total float64) (model.OutgoingInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.OutgoingInvoice{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
UPDATE outgoing_invoices
SET
  invoice_date = to_timestamp($2::double precision / 1000.0),
  deal_id = $3::uuid,
  total = $4,
  comment = $5,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  invoice_number,
  invoice_date,
  deal_id::text,
  total,
  comment,
  created_at
`
	invoice, err := scanOutgoingInvoice(tx.QueryRow(ctx, query, id, input.Date, input.DealID, total, input.Comment))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.OutgoingInvoice{}, ErrOutgoingInvoiceNotFound
		}
		if isUniqueViolation(err) {
			return model.OutgoingInvoice{}, ErrOutgoingInvoiceDealTaken
		}
		if isForeignKeyViolation(err) {
			return model.OutgoingInvoice{}, errors.New("сделка не найдена")
		}
		return model.OutgoingInvoice{}, err
	}

	if err := replaceInvoiceItems(ctx, tx, "outgoing", invoice.ID, input.Items); err != nil {
		return model.OutgoingInvoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.OutgoingInvoice{}, err
	}

	invoice.Items = append([]model.InvoiceItem(nil), input.Items...)
	return invoice, nil
}

func (r *OutgoingInvoiceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM outgoing_invoices WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOutgoingInvoiceNotFound
	}
	return nil
}

type outgoingInvoiceScanner interface {
	Scan(dest ...any) error
}

func scanOutgoingInvoice(row outgoingInvoiceScanner) (model.OutgoingInvoice, error) {
	var item model.OutgoingInvoice
	var invoiceDate time.Time
	var createdAt time.Time
	err := row.Scan(
		&item.ID,
		&item.InvoiceNumber,
		&invoiceDate,
		&item.DealID,
		&item.Total,
		&item.Comment,
		&createdAt,
	)
	if err != nil {
		return model.OutgoingInvoice{}, err
	}
	item.Date = invoiceDate.UnixMilli()
	item.CreatedAt = createdAt.UnixMilli()
	if item.Items == nil {
		item.Items = make([]model.InvoiceItem, 0)
	}
	return item, nil
}
