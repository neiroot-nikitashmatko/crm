package repository

import (
	"context"
	"errors"
	"time"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAttachmentNotFound = errors.New("attachment not found")

type AttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(
	ctx context.Context,
	entityType string,
	entityID string,
	uploadedBy string,
	name string,
	mimeType string,
	content []byte,
) (model.Attachment, error) {
	const query = `
INSERT INTO attachments (
  entity_type,
  entity_id,
  name,
  size,
  mime_type,
  content,
  uploaded_by
)
VALUES (
  $1,
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7::uuid
)
RETURNING
  id::text,
  name,
  size,
  mime_type,
  uploaded_by::text,
  created_at
`
	row := r.db.QueryRow(ctx, query, entityType, entityID, name, int64(len(content)), mimeType, content, uploadedBy)
	return scanAttachmentMeta(row)
}

func (r *AttachmentRepository) ListByEntity(ctx context.Context, entityType string, entityID string) ([]model.Attachment, error) {
	const query = `
SELECT
  a.id::text,
  a.name,
  a.size,
  a.mime_type,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, a.uploaded_by::text) AS uploaded_by_label,
  a.created_at
FROM attachments a
LEFT JOIN users u ON u.id = a.uploaded_by
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

	result := make([]model.Attachment, 0)
	for rows.Next() {
		item, err := scanAttachmentMetaRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *AttachmentRepository) ListByEntityIDs(
	ctx context.Context,
	entityType string,
	entityIDs []string,
) (map[string][]model.Attachment, error) {
	result := make(map[string][]model.Attachment)
	if len(entityIDs) == 0 {
		return result, nil
	}

	const query = `
SELECT
  a.entity_id::text,
  a.id::text,
  a.name,
  a.size,
  a.mime_type,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, a.uploaded_by::text) AS uploaded_by_label,
  a.created_at
FROM attachments a
LEFT JOIN users u ON u.id = a.uploaded_by
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
		var item model.Attachment
		if err := rows.Scan(
			&entityID,
			&item.ID,
			&item.Name,
			&item.Size,
			&item.MimeType,
			&item.UploadedBy,
			&createdAt,
		); err != nil {
			return nil, err
		}
		item.UploadedAt = createdAt.UnixMilli()
		result[entityID] = append(result[entityID], item)
	}
	return result, rows.Err()
}

type AttachmentMeta struct {
	ID         string
	Name       string
	EntityType string
	EntityID   string
}

func (r *AttachmentRepository) GetMeta(ctx context.Context, attachmentID string) (AttachmentMeta, error) {
	const query = `
SELECT
  a.id::text,
  a.name,
  a.entity_type,
  a.entity_id::text
FROM attachments a
WHERE a.id = $1::uuid
  AND a.deleted_at IS NULL
LIMIT 1
`
	var item AttachmentMeta
	err := r.db.QueryRow(ctx, query, attachmentID).Scan(
		&item.ID,
		&item.Name,
		&item.EntityType,
		&item.EntityID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AttachmentMeta{}, ErrAttachmentNotFound
		}
		return AttachmentMeta{}, err
	}
	return item, nil
}

func (r *AttachmentRepository) GetContent(ctx context.Context, attachmentID string) (model.AttachmentContent, error) {
	const query = `
SELECT
  a.id::text,
  a.name,
  a.size,
  a.mime_type,
  COALESCE(NULLIF(trim(concat_ws(' ', u.last_name, u.first_name, u.patronymic)), ''), u.phone, a.uploaded_by::text) AS uploaded_by_label,
  a.created_at,
  a.content
FROM attachments a
LEFT JOIN users u ON u.id = a.uploaded_by
WHERE a.id = $1::uuid
  AND a.deleted_at IS NULL
LIMIT 1
`
	var createdAt time.Time
	var item model.AttachmentContent
	err := r.db.QueryRow(ctx, query, attachmentID).Scan(
		&item.ID,
		&item.Name,
		&item.Size,
		&item.MimeType,
		&item.UploadedBy,
		&createdAt,
		&item.Content,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AttachmentContent{}, ErrAttachmentNotFound
		}
		return model.AttachmentContent{}, err
	}
	item.UploadedAt = createdAt.UnixMilli()
	return item, nil
}

func (r *AttachmentRepository) SoftDelete(ctx context.Context, attachmentID string) error {
	const query = `
UPDATE attachments
SET deleted_at = now()
WHERE id = $1::uuid
  AND deleted_at IS NULL
`
	tag, err := r.db.Exec(ctx, query, attachmentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (r *AttachmentRepository) DealExists(ctx context.Context, dealID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM deals WHERE id = $1::uuid AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, dealID).Scan(&exists)
	return exists, err
}

func (r *AttachmentRepository) TaskExists(ctx context.Context, taskID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1::uuid)`
	var exists bool
	err := r.db.QueryRow(ctx, query, taskID).Scan(&exists)
	return exists, err
}

type attachmentMetaScanner interface {
	Scan(dest ...any) error
}

func scanAttachmentMeta(row attachmentMetaScanner) (model.Attachment, error) {
	var item model.Attachment
	var createdAt time.Time
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Size,
		&item.MimeType,
		&item.UploadedBy,
		&createdAt,
	)
	if err != nil {
		return model.Attachment{}, err
	}
	item.UploadedAt = createdAt.UnixMilli()
	return item, nil
}

func scanAttachmentMetaRow(row attachmentMetaScanner) (model.Attachment, error) {
	return scanAttachmentMeta(row)
}
