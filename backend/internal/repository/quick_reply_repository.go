package repository

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuickReplyRepository struct {
	db *pgxpool.Pool
}

func NewQuickReplyRepository(db *pgxpool.Pool) *QuickReplyRepository {
	return &QuickReplyRepository{db: db}
}

func (r *QuickReplyRepository) ListSectionsWithReplies(ctx context.Context) ([]model.QuickReplySection, error) {
	sectionsQuery := `
SELECT id::text, title, sort_order, created_at, updated_at
FROM quick_reply_sections
ORDER BY sort_order ASC, created_at ASC
`
	rows, err := r.db.Query(ctx, sectionsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := make([]model.QuickReplySection, 0)
	indexByID := make(map[string]int)
	for rows.Next() {
		var section model.QuickReplySection
		if scanErr := rows.Scan(
			&section.ID,
			&section.Title,
			&section.SortOrder,
			&section.CreatedAt,
			&section.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		section.Replies = []model.QuickReply{}
		indexByID[section.ID] = len(sections)
		sections = append(sections, section)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		return sections, nil
	}

	repliesQuery := `
SELECT id::text, section_id::text, title, body, sort_order, created_at, updated_at
FROM quick_replies
ORDER BY sort_order ASC, created_at ASC
`
	replyRows, err := r.db.Query(ctx, repliesQuery)
	if err != nil {
		return nil, err
	}
	defer replyRows.Close()

	for replyRows.Next() {
		var reply model.QuickReply
		if scanErr := replyRows.Scan(
			&reply.ID,
			&reply.SectionID,
			&reply.Title,
			&reply.Body,
			&reply.SortOrder,
			&reply.CreatedAt,
			&reply.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}
		idx, ok := indexByID[reply.SectionID]
		if !ok {
			continue
		}
		sections[idx].Replies = append(sections[idx].Replies, reply)
	}
	return sections, replyRows.Err()
}

func (r *QuickReplyRepository) CreateSection(ctx context.Context, title string) (model.QuickReplySection, error) {
	query := `
INSERT INTO quick_reply_sections (title, sort_order)
VALUES (
  $1,
  COALESCE((SELECT MAX(sort_order) + 1 FROM quick_reply_sections), 0)
)
RETURNING id::text, title, sort_order, created_at, updated_at
`
	var section model.QuickReplySection
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(title)).Scan(
		&section.ID,
		&section.Title,
		&section.SortOrder,
		&section.CreatedAt,
		&section.UpdatedAt,
	)
	section.Replies = []model.QuickReply{}
	return section, err
}

func (r *QuickReplyRepository) UpdateSection(ctx context.Context, id, title string) (model.QuickReplySection, error) {
	query := `
UPDATE quick_reply_sections
SET title = $2, updated_at = now()
WHERE id = $1::uuid
RETURNING id::text, title, sort_order, created_at, updated_at
`
	var section model.QuickReplySection
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(id), strings.TrimSpace(title)).Scan(
		&section.ID,
		&section.Title,
		&section.SortOrder,
		&section.CreatedAt,
		&section.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.QuickReplySection{}, pgx.ErrNoRows
	}
	section.Replies = []model.QuickReply{}
	return section, err
}

func (r *QuickReplyRepository) DeleteSection(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM quick_reply_sections WHERE id = $1::uuid`, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *QuickReplyRepository) CreateReply(ctx context.Context, sectionID, title, body string) (model.QuickReply, error) {
	query := `
INSERT INTO quick_replies (section_id, title, body, sort_order)
VALUES (
  $1::uuid,
  $2,
  $3,
  COALESCE((SELECT MAX(sort_order) + 1 FROM quick_replies WHERE section_id = $1::uuid), 0)
)
RETURNING id::text, section_id::text, title, body, sort_order, created_at, updated_at
`
	var reply model.QuickReply
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(sectionID), strings.TrimSpace(title), strings.TrimSpace(body)).Scan(
		&reply.ID,
		&reply.SectionID,
		&reply.Title,
		&reply.Body,
		&reply.SortOrder,
		&reply.CreatedAt,
		&reply.UpdatedAt,
	)
	return reply, err
}

func (r *QuickReplyRepository) UpdateReply(ctx context.Context, id, title, body string) (model.QuickReply, error) {
	query := `
UPDATE quick_replies
SET title = $2, body = $3, updated_at = now()
WHERE id = $1::uuid
RETURNING id::text, section_id::text, title, body, sort_order, created_at, updated_at
`
	var reply model.QuickReply
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(id), strings.TrimSpace(title), strings.TrimSpace(body)).Scan(
		&reply.ID,
		&reply.SectionID,
		&reply.Title,
		&reply.Body,
		&reply.SortOrder,
		&reply.CreatedAt,
		&reply.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.QuickReply{}, pgx.ErrNoRows
	}
	return reply, err
}

func (r *QuickReplyRepository) DeleteReply(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM quick_replies WHERE id = $1::uuid`, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *QuickReplyRepository) SectionExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM quick_reply_sections WHERE id = $1::uuid)`, strings.TrimSpace(id)).Scan(&exists)
	return exists, err
}
