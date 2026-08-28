package repository

import (
	"context"
	"errors"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSupplierNotFound = errors.New("поставщик не найден")

type SupplierRepository struct {
	db *pgxpool.Pool
}

func NewSupplierRepository(db *pgxpool.Pool) *SupplierRepository {
	return &SupplierRepository{db: db}
}

const supplierSelect = `
SELECT
  id::text,
  name,
  contact_person,
  phone,
  inn,
  kpp,
  ogrn,
  legal_address,
  actual_address,
  bik,
  settlement_account,
  correspondent_account,
  created_at
FROM suppliers
`

func (r *SupplierRepository) List(ctx context.Context) ([]model.Supplier, error) {
	query := supplierSelect + `
ORDER BY name ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Supplier, 0)
	for rows.Next() {
		item, err := scanSupplier(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SupplierRepository) GetByID(ctx context.Context, id string) (model.Supplier, error) {
	query := supplierSelect + `
WHERE id = $1::uuid
`
	item, err := scanSupplier(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supplier{}, ErrSupplierNotFound
		}
		return model.Supplier{}, err
	}
	return item, nil
}

func (r *SupplierRepository) Create(ctx context.Context, input model.UpsertSupplierInput) (model.Supplier, error) {
	const query = `
INSERT INTO suppliers (
  name,
  contact_person,
  phone,
  inn,
  kpp,
  ogrn,
  legal_address,
  actual_address,
  bik,
  settlement_account,
  correspondent_account
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
  id::text,
  name,
  contact_person,
  phone,
  inn,
  kpp,
  ogrn,
  legal_address,
  actual_address,
  bik,
  settlement_account,
  correspondent_account,
  created_at
`
	return scanSupplier(r.db.QueryRow(
		ctx,
		query,
		input.Name,
		input.ContactPerson,
		input.Phone,
		input.INN,
		input.KPP,
		input.OGRN,
		input.LegalAddress,
		input.ActualAddress,
		input.BIK,
		input.SettlementAccount,
		input.CorrespondentAccount,
	))
}

func (r *SupplierRepository) Update(ctx context.Context, id string, input model.UpsertSupplierInput) (model.Supplier, error) {
	const query = `
UPDATE suppliers
SET
  name = $2,
  contact_person = $3,
  phone = $4,
  inn = $5,
  kpp = $6,
  ogrn = $7,
  legal_address = $8,
  actual_address = $9,
  bik = $10,
  settlement_account = $11,
  correspondent_account = $12,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  name,
  contact_person,
  phone,
  inn,
  kpp,
  ogrn,
  legal_address,
  actual_address,
  bik,
  settlement_account,
  correspondent_account,
  created_at
`
	item, err := scanSupplier(r.db.QueryRow(
		ctx,
		query,
		id,
		input.Name,
		input.ContactPerson,
		input.Phone,
		input.INN,
		input.KPP,
		input.OGRN,
		input.LegalAddress,
		input.ActualAddress,
		input.BIK,
		input.SettlementAccount,
		input.CorrespondentAccount,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Supplier{}, ErrSupplierNotFound
		}
		return model.Supplier{}, err
	}
	return item, nil
}

func (r *SupplierRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM suppliers WHERE id = $1::uuid`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return errors.New("нельзя удалить поставщика: есть связанные приходные накладные")
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSupplierNotFound
	}
	return nil
}

type supplierScanner interface {
	Scan(dest ...any) error
}

func scanSupplier(row supplierScanner) (model.Supplier, error) {
	var item model.Supplier
	var createdAt time.Time
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.ContactPerson,
		&item.Phone,
		&item.INN,
		&item.KPP,
		&item.OGRN,
		&item.LegalAddress,
		&item.ActualAddress,
		&item.BIK,
		&item.SettlementAccount,
		&item.CorrespondentAccount,
		&createdAt,
	)
	if err != nil {
		return model.Supplier{}, err
	}
	item.CreatedAt = createdAt.UnixMilli()
	return item, nil
}
