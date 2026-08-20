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

var ErrPaymentNotFound = errors.New("оплата не найдена")

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

const paymentSelect = `
SELECT
  id::text,
  payment_date,
  remind_at,
  payer_id::text,
  counterparty,
  amount,
  short_title,
  comment,
  is_closed,
  created_by::text,
  created_at,
  updated_at
FROM payments
`

func (r *PaymentRepository) List(ctx context.Context) ([]model.Payment, error) {
	query := paymentSelect + `
ORDER BY payment_date ASC, created_at ASC
`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Payment, 0)
	for rows.Next() {
		item, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PaymentRepository) Create(ctx context.Context, createdBy string, input model.CreatePaymentInput) (model.Payment, error) {
	const query = `
INSERT INTO payments (
  short_title,
  payment_date,
  remind_at,
  payer_id,
  counterparty,
  amount,
  comment,
  created_by
) VALUES (
  $1,
  to_timestamp($2::double precision / 1000.0),
  CASE
    WHEN $3::double precision IS NULL THEN NULL
    ELSE to_timestamp($3::double precision / 1000.0)
  END,
  NULLIF($4::text, '')::payment_payer,
  $5,
  $6,
  $7,
  $8::uuid
)
RETURNING
  id::text,
  payment_date,
  remind_at,
  payer_id::text,
  counterparty,
  amount,
  short_title,
  comment,
  is_closed,
  created_by::text,
  created_at,
  updated_at
`
	row := r.db.QueryRow(
		ctx,
		query,
		input.ShortTitle,
		input.Date,
		optionalMillis(input.RemindAt),
		optionalPayer(input.PayerID),
		input.Counterparty,
		input.Amount,
		input.Comment,
		createdBy,
	)
	item, err := scanPayment(row)
	if err != nil {
		return model.Payment{}, mapPaymentWriteError(err)
	}
	return item, nil
}

func (r *PaymentRepository) Update(ctx context.Context, id string, input model.CreatePaymentInput) (model.Payment, error) {
	const query = `
UPDATE payments
SET
  short_title = $2,
  payment_date = to_timestamp($3::double precision / 1000.0),
  remind_at = CASE
    WHEN $4::double precision IS NULL THEN NULL
    ELSE to_timestamp($4::double precision / 1000.0)
  END,
  payer_id = NULLIF($5::text, '')::payment_payer,
  counterparty = $6,
  amount = $7,
  comment = $8,
  reminder_sent_at = CASE
    WHEN remind_at IS NOT DISTINCT FROM (
      CASE
        WHEN $4::double precision IS NULL THEN NULL
        ELSE to_timestamp($4::double precision / 1000.0)
      END
    ) THEN reminder_sent_at
    ELSE NULL
  END,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  payment_date,
  remind_at,
  payer_id::text,
  counterparty,
  amount,
  short_title,
  comment,
  is_closed,
  created_by::text,
  created_at,
  updated_at
`
	row := r.db.QueryRow(
		ctx,
		query,
		id,
		input.ShortTitle,
		input.Date,
		optionalMillis(input.RemindAt),
		optionalPayer(input.PayerID),
		input.Counterparty,
		input.Amount,
		input.Comment,
	)
	item, err := scanPayment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, ErrPaymentNotFound
		}
		return model.Payment{}, mapPaymentWriteError(err)
	}
	return item, nil
}

func (r *PaymentRepository) SetClosed(ctx context.Context, id string, isClosed bool) (model.Payment, error) {
	const query = `
UPDATE payments
SET
  is_closed = $2,
  closed_at = CASE
    WHEN $2 THEN COALESCE(closed_at, now())
    ELSE NULL
  END,
  updated_at = now()
WHERE id = $1::uuid
RETURNING
  id::text,
  payment_date,
  remind_at,
  payer_id::text,
  counterparty,
  amount,
  short_title,
  comment,
  is_closed,
  created_by::text,
  created_at,
  updated_at
`
	row := r.db.QueryRow(ctx, query, id, isClosed)
	item, err := scanPayment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Payment{}, ErrPaymentNotFound
		}
		return model.Payment{}, mapPaymentWriteError(err)
	}
	return item, nil
}

func (r *PaymentRepository) ListDueReminders(ctx context.Context, moscowDate time.Time) ([]model.Payment, error) {
	query := paymentSelect + `
WHERE remind_at IS NOT NULL
  AND reminder_sent_at IS NULL
  AND is_closed = FALSE
  AND (timezone('Europe/Moscow', remind_at))::date <= $1::date
ORDER BY remind_at ASC, created_at ASC
`
	rows, err := r.db.Query(ctx, query, moscowDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Payment, 0)
	for rows.Next() {
		item, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *PaymentRepository) MarkReminderSent(ctx context.Context, id string) error {
	const query = `
UPDATE payments
SET
  reminder_sent_at = now(),
  updated_at = now()
WHERE id = $1::uuid
  AND reminder_sent_at IS NULL
  AND is_closed = FALSE
  AND remind_at IS NOT NULL
`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PaymentRepository) Delete(ctx context.Context, id string) error {
	const query = `
DELETE FROM payments
WHERE id = $1::uuid
`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

type paymentScanner interface {
	Scan(dest ...any) error
}

func scanPayment(row paymentScanner) (model.Payment, error) {
	var item model.Payment
	var paymentDate time.Time
	var remindAt *time.Time
	var payerID *string
	var createdAt time.Time
	var updatedAt time.Time

	if err := row.Scan(
		&item.ID,
		&paymentDate,
		&remindAt,
		&payerID,
		&item.Counterparty,
		&item.Amount,
		&item.ShortTitle,
		&item.Comment,
		&item.IsClosed,
		&item.CreatedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.Payment{}, err
	}

	item.Date = paymentDate.UnixMilli()
	if remindAt != nil {
		ms := remindAt.UnixMilli()
		item.RemindAt = &ms
	}
	item.PayerID = payerID
	item.CreatedAt = createdAt.UnixMilli()
	item.UpdatedAt = updatedAt.UnixMilli()
	return item, nil
}

func optionalMillis(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalPayer(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func mapPaymentWriteError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "payments_remind_not_after_payment") {
		return errors.New("дата напоминания не может быть позже даты платежа")
	}
	if strings.Contains(msg, "invalid input value for enum payment_payer") {
		return errors.New("некорректный плательщик")
	}
	if strings.Contains(msg, "payments_short_title_check") {
		return errors.New("краткое описание обязательно и не длиннее 22 символов")
	}
	if strings.Contains(msg, "payments_created_by_fkey") {
		return errors.New("authorization required")
	}
	return fmt.Errorf("%w", err)
}
