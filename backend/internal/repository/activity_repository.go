package repository

import (
	"context"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepository struct {
	db *pgxpool.Pool
}

func NewActivityRepository(db *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(
	ctx context.Context,
	entityType string,
	entityID string,
	authorID string,
	activityType string,
	text string,
) (model.Activity, error) {
	const query = `
WITH inserted AS (
  INSERT INTO activities (
    entity_type,
    entity_id,
    activity_type,
    text,
    author_id
  )
  VALUES (
    $1,
    $2::uuid,
    $3,
    $4,
    $5::uuid
  )
  RETURNING id, activity_type, text, author_id, created_at
)
SELECT
  i.id::text,
  i.activity_type,
  i.text,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, i.author_id::text) AS author_label,
  i.created_at
FROM inserted i
LEFT JOIN users u ON u.id = i.author_id
`
	row := r.db.QueryRow(ctx, query, entityType, entityID, activityType, text, authorID)
	return scanActivityMeta(row)
}

func (r *ActivityRepository) ListByEntity(ctx context.Context, entityType string, entityID string) ([]model.Activity, error) {
	const query = `
SELECT
  a.id::text,
  a.activity_type,
  a.text,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, a.author_id::text) AS author_label,
  a.created_at
FROM activities a
LEFT JOIN users u ON u.id = a.author_id
WHERE a.entity_type = $1
  AND a.entity_id = $2::uuid
  AND a.deleted_at IS NULL
ORDER BY a.created_at DESC
`
	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.Activity, 0)
	for rows.Next() {
		item, err := scanActivityMetaRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ActivityRepository) ListByEntityIDs(
	ctx context.Context,
	entityType string,
	entityIDs []string,
) (map[string][]model.Activity, error) {
	result := make(map[string][]model.Activity)
	if len(entityIDs) == 0 {
		return result, nil
	}

	const query = `
SELECT
  a.entity_id::text,
  a.id::text,
  a.activity_type,
  a.text,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, a.author_id::text) AS author_label,
  a.created_at
FROM activities a
LEFT JOIN users u ON u.id = a.author_id
WHERE a.entity_type = $1
  AND a.entity_id = ANY($2::uuid[])
  AND a.deleted_at IS NULL
ORDER BY a.created_at DESC
`
	rows, err := r.db.Query(ctx, query, entityType, entityIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID string
		var createdAt time.Time
		var item model.Activity
		if err := rows.Scan(
			&entityID,
			&item.ID,
			&item.Type,
			&item.Text,
			&item.Author,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.UnixMilli()
		result[entityID] = append(result[entityID], item)
	}
	return result, rows.Err()
}

func (r *ActivityRepository) DealExists(ctx context.Context, dealID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM deals WHERE id = $1::uuid AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, dealID).Scan(&exists)
	return exists, err
}

func (r *ActivityRepository) LeadExists(ctx context.Context, leadID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM leads WHERE id = $1::uuid AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, leadID).Scan(&exists)
	return exists, err
}

func (r *ActivityRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1::uuid)`
	var exists bool
	err := r.db.QueryRow(ctx, query, taskID).Scan(&exists)
	return exists, err
}

type activityMetaScanner interface {
	Scan(dest ...any) error
}

func scanActivityMeta(row activityMetaScanner) (model.Activity, error) {
	var item model.Activity
	var createdAt time.Time
	err := row.Scan(
		&item.ID,
		&item.Type,
		&item.Text,
		&item.Author,
		&createdAt,
	)
	if err != nil {
		return model.Activity{}, err
	}
	item.CreatedAt = createdAt.UnixMilli()
	return item, nil
}

func scanActivityMetaRow(row activityMetaScanner) (model.Activity, error) {
	return scanActivityMeta(row)
}
