package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SalaryEntryRepository struct {
	db *pgxpool.Pool
}

func NewSalaryEntryRepository(db *pgxpool.Pool) *SalaryEntryRepository {
	return &SalaryEntryRepository{db: db}
}

const salaryEntrySelect = `
SELECT
  se.id::text,
  se.entry_date,
  se.deal_id::text,
  '#' || d.deal_number::text,
  se.service::text,
  se.salary_amount,
  se.comment,
  se.created_by::text,
  se.created_at,
  se.updated_at
FROM salary_entries se
JOIN deals d ON d.id = se.deal_id
`

func (r *SalaryEntryRepository) ListByCreatedBy(ctx context.Context, createdBy string) ([]model.SalaryEntry, error) {
	query := salaryEntrySelect + `
WHERE se.created_by = $1::uuid
  AND d.deleted_at IS NULL
ORDER BY se.entry_date DESC, se.created_at DESC
`
	return r.querySalaryEntries(ctx, query, createdBy)
}

func (r *SalaryEntryRepository) ListAll(ctx context.Context) ([]model.SalaryEntry, error) {
	query := salaryEntrySelect + `
WHERE d.deleted_at IS NULL
ORDER BY se.entry_date DESC, se.created_at DESC
`
	return r.querySalaryEntries(ctx, query)
}

func (r *SalaryEntryRepository) querySalaryEntries(ctx context.Context, query string, args ...any) ([]model.SalaryEntry, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.SalaryEntry, 0)
	for rows.Next() {
		item, err := scanSalaryEntry(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *SalaryEntryRepository) GetByIDForUser(ctx context.Context, id string, createdBy string) (model.SalaryEntry, error) {
	query := salaryEntrySelect + `
WHERE se.id = $1::uuid
  AND se.created_by = $2::uuid
  AND d.deleted_at IS NULL
`
	row := r.db.QueryRow(ctx, query, id, createdBy)
	item, err := scanSalaryEntry(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SalaryEntry{}, errors.New("запись не найдена")
		}
		return model.SalaryEntry{}, err
	}
	return item, nil
}

func (r *SalaryEntryRepository) GetByID(ctx context.Context, id string) (model.SalaryEntry, error) {
	query := salaryEntrySelect + `
WHERE se.id = $1::uuid
  AND d.deleted_at IS NULL
`
	row := r.db.QueryRow(ctx, query, id)
	item, err := scanSalaryEntry(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SalaryEntry{}, errors.New("запись не найдена")
		}
		return model.SalaryEntry{}, err
	}
	return item, nil
}

func (r *SalaryEntryRepository) Create(ctx context.Context, createdBy string, input model.UpsertSalaryEntryInput) (model.SalaryEntry, error) {
	const query = `
WITH inserted AS (
  INSERT INTO salary_entries (entry_date, deal_id, service, salary_amount, comment, created_by)
  VALUES (to_timestamp($1::double precision / 1000.0), $2::uuid, $3::salary_service, $4, $5, $6::uuid)
  RETURNING *
)
SELECT
  inserted.id::text,
  inserted.entry_date,
  inserted.deal_id::text,
  '#' || d.deal_number::text,
  inserted.service::text,
  inserted.salary_amount,
  inserted.comment,
  inserted.created_by::text,
  inserted.created_at,
  inserted.updated_at
FROM inserted
JOIN deals d ON d.id = inserted.deal_id
`
	row := r.db.QueryRow(
		ctx,
		query,
		input.Date,
		input.DealID,
		input.Service,
		input.Salary,
		input.Comment,
		createdBy,
	)
	item, err := scanSalaryEntry(row)
	if err != nil {
		return model.SalaryEntry{}, mapSalaryEntryWriteError(err)
	}
	return item, nil
}

func (r *SalaryEntryRepository) Update(ctx context.Context, id string, createdBy string, input model.UpsertSalaryEntryInput) (model.SalaryEntry, error) {
	const query = `
WITH updated AS (
  UPDATE salary_entries
  SET
    entry_date = to_timestamp($3::double precision / 1000.0),
    deal_id = $4::uuid,
    service = $5::salary_service,
    salary_amount = $6,
    comment = $7,
    updated_at = now()
  WHERE id = $1::uuid
    AND created_by = $2::uuid
  RETURNING *
)
SELECT
  updated.id::text,
  updated.entry_date,
  updated.deal_id::text,
  '#' || d.deal_number::text,
  updated.service::text,
  updated.salary_amount,
  updated.comment,
  updated.created_by::text,
  updated.created_at,
  updated.updated_at
FROM updated
JOIN deals d ON d.id = updated.deal_id
`
	row := r.db.QueryRow(
		ctx,
		query,
		id,
		createdBy,
		input.Date,
		input.DealID,
		input.Service,
		input.Salary,
		input.Comment,
	)
	item, err := scanSalaryEntry(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SalaryEntry{}, errors.New("запись не найдена")
		}
		return model.SalaryEntry{}, mapSalaryEntryWriteError(err)
	}
	return item, nil
}

func (r *SalaryEntryRepository) UpdateByID(ctx context.Context, id string, input model.UpsertSalaryEntryInput) (model.SalaryEntry, error) {
	const query = `
WITH updated AS (
  UPDATE salary_entries
  SET
    entry_date = to_timestamp($2::double precision / 1000.0),
    deal_id = $3::uuid,
    service = $4::salary_service,
    salary_amount = $5,
    comment = $6,
    updated_at = now()
  WHERE id = $1::uuid
  RETURNING *
)
SELECT
  updated.id::text,
  updated.entry_date,
  updated.deal_id::text,
  '#' || d.deal_number::text,
  updated.service::text,
  updated.salary_amount,
  updated.comment,
  updated.created_by::text,
  updated.created_at,
  updated.updated_at
FROM updated
JOIN deals d ON d.id = updated.deal_id
`
	row := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Date,
		input.DealID,
		input.Service,
		input.Salary,
		input.Comment,
	)
	item, err := scanSalaryEntry(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SalaryEntry{}, errors.New("запись не найдена")
		}
		return model.SalaryEntry{}, mapSalaryEntryWriteError(err)
	}
	return item, nil
}

func (r *SalaryEntryRepository) Delete(ctx context.Context, id string, createdBy string) error {
	const query = `
DELETE FROM salary_entries
WHERE id = $1::uuid
  AND created_by = $2::uuid
`
	tag, err := r.db.Exec(ctx, query, id, createdBy)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("запись не найдена")
	}
	return nil
}

func (r *SalaryEntryRepository) DeleteByID(ctx context.Context, id string) error {
	const query = `
DELETE FROM salary_entries
WHERE id = $1::uuid
`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("запись не найдена")
	}
	return nil
}

type salaryEntryScanner interface {
	Scan(dest ...any) error
}

func scanSalaryEntry(row salaryEntryScanner) (model.SalaryEntry, error) {
	var item model.SalaryEntry
	var entryDate time.Time
	var createdAt time.Time
	var updatedAt time.Time

	if err := row.Scan(
		&item.ID,
		&entryDate,
		&item.DealID,
		&item.DealNumberLabel,
		&item.Service,
		&item.Salary,
		&item.Comment,
		&item.CreatedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.SalaryEntry{}, err
	}

	item.Date = entryDate.UnixMilli()
	item.CreatedAt = createdAt.UnixMilli()
	item.UpdatedAt = updatedAt.UnixMilli()
	return item, nil
}

func mapSalaryEntryWriteError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "salary_entries_deal_id_fkey") {
		return errors.New("сделка не найдена")
	}
	if strings.Contains(msg, "invalid input value for enum salary_service") {
		return errors.New("некорректная услуга")
	}
	return fmt.Errorf("%w", err)
}
