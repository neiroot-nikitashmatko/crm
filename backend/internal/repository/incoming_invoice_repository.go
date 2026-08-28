package repository

import (
	"context"
	"errors"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrIncomingInvoiceNotFound = errors.New("приходная накладная не найдена")

type IncomingInvoiceRepository struct {
	db *pgxpool.Pool
}

func NewIncomingInvoiceRepository(db *pgxpool.Pool) *IncomingInvoiceRepository {
	return &IncomingInvoiceRepository{db: db}
}

const incomingInvoiceSelect = `
SELECT
  id::text,
  invoice_number,
  invoice_date,
  supplier_id::text,
  total,
  comment,
  created_at
FROM incoming_invoices
`

func (r *IncomingInvoiceRepository) List(ctx context.Context) ([]model.IncomingInvoice, error) {
	query := incomingInvoiceSelect + `
ORDER BY invoice_date DESC, created_at DESC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]model.IncomingInvoice, 0)
	for rows.Next() {
		item, err := scanIncomingInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	itemsByInvoice, err := listInvoiceItemsByKind(ctx, r.db, "incoming")
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

func (r *IncomingInvoiceRepository) Create(ctx context.Context, input model.UpsertIncomingInvoiceInput, total float64) (model.IncomingInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.IncomingInvoice{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
INSERT INTO incoming_invoices (invoice_date, supplier_id, total, comment)
VALUES (to_timestamp($1::double precision / 1000.0), $2::uuid, $3, $4)
RETURNING
  id::text,
  invoice_number,
  invoice_date,
  supplier_id::text,
  total,
  comment,
  created_at
`
	invoice, err := scanIncomingInvoice(tx.QueryRow(ctx, query, input.Date, input.SupplierID, total, input.Comment))
	if err != nil {
		if isForeignKeyViolation(err) {
			return model.IncomingInvoice{}, errors.New("поставщик не найден")
		}
		return model.IncomingInvoice{}, err
	}

	if err := replaceInvoiceItems(ctx, tx, "incoming", invoice.ID, input.Items); err != nil {
		return model.IncomingInvoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.IncomingInvoice{}, err
	}

	invoice.Items = append([]model.InvoiceItem(nil), input.Items...)
	return invoice, nil
}

func (r *IncomingInvoiceRepository) Update(ctx context.Context, id string, input model.UpsertIncomingInvoiceInput, total float64) (model.IncomingInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return model.IncomingInvoice{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const query = `
UPDATE incoming_invoices
SET
  invoice_date = to_timestamp($2::double precision / 1000.0),
  supplier_id = $3::uuid,
  total = $4,
  comment = $5,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  invoice_number,
  invoice_date,
  supplier_id::text,
  total,
  comment,
  created_at
`
	invoice, err := scanIncomingInvoice(tx.QueryRow(ctx, query, id, input.Date, input.SupplierID, total, input.Comment))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.IncomingInvoice{}, ErrIncomingInvoiceNotFound
		}
		if isForeignKeyViolation(err) {
			return model.IncomingInvoice{}, errors.New("поставщик не найден")
		}
		return model.IncomingInvoice{}, err
	}

	if err := replaceInvoiceItems(ctx, tx, "incoming", invoice.ID, input.Items); err != nil {
		return model.IncomingInvoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.IncomingInvoice{}, err
	}

	invoice.Items = append([]model.InvoiceItem(nil), input.Items...)
	return invoice, nil
}

func (r *IncomingInvoiceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM incoming_invoices WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrIncomingInvoiceNotFound
	}
	return nil
}

type incomingInvoiceScanner interface {
	Scan(dest ...any) error
}

func scanIncomingInvoice(row incomingInvoiceScanner) (model.IncomingInvoice, error) {
	var item model.IncomingInvoice
	var invoiceDate time.Time
	var createdAt time.Time
	err := row.Scan(
		&item.ID,
		&item.InvoiceNumber,
		&invoiceDate,
		&item.SupplierID,
		&item.Total,
		&item.Comment,
		&createdAt,
	)
	if err != nil {
		return model.IncomingInvoice{}, err
	}
	item.Date = invoiceDate.UnixMilli()
	item.CreatedAt = createdAt.UnixMilli()
	if item.Items == nil {
		item.Items = make([]model.InvoiceItem, 0)
	}
	return item, nil
}
