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

type AvitoChatRepository struct {
	db *pgxpool.Pool
}

func NewAvitoChatRepository(db *pgxpool.Pool) *AvitoChatRepository {
	return &AvitoChatRepository{db: db}
}

func (r *AvitoChatRepository) UpsertChat(ctx context.Context, input model.UpsertAvitoChatInput) (model.AvitoChat, error) {
	query := `
INSERT INTO avito_chats (
  chat_id, lead_id, peer_user_id, peer_nickname, peer_avatar_url, item_id, item_title
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (chat_id) DO UPDATE SET
  lead_id = EXCLUDED.lead_id,
  peer_user_id = EXCLUDED.peer_user_id,
  peer_nickname = EXCLUDED.peer_nickname,
  peer_avatar_url = EXCLUDED.peer_avatar_url,
  item_id = EXCLUDED.item_id,
  item_title = EXCLUDED.item_title,
  updated_at = now()
RETURNING
  id::text,
  chat_id,
  lead_id::text,
  peer_user_id,
  peer_nickname,
  peer_avatar_url,
  item_id,
  item_title,
  created_at,
  updated_at
`
	row := r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.ChatID),
		strings.TrimSpace(input.LeadID),
		input.PeerUserID,
		strings.TrimSpace(input.PeerNickname),
		strings.TrimSpace(input.PeerAvatarURL),
		input.ItemID,
		strings.TrimSpace(input.ItemTitle),
	)
	return scanAvitoChat(row)
}

func (r *AvitoChatRepository) GetByChatID(ctx context.Context, chatID string) (model.AvitoChat, error) {
	query := `
SELECT
  id::text,
  chat_id,
  lead_id::text,
  peer_user_id,
  peer_nickname,
  peer_avatar_url,
  item_id,
  item_title,
  created_at,
  updated_at
FROM avito_chats
WHERE chat_id = $1
`
	chat, err := scanAvitoChat(r.db.QueryRow(ctx, query, strings.TrimSpace(chatID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AvitoChat{}, pgx.ErrNoRows
	}
	return chat, err
}

func (r *AvitoChatRepository) GetByLeadID(ctx context.Context, leadID string) (model.AvitoChat, error) {
	query := `
SELECT
  id::text,
  chat_id,
  lead_id::text,
  peer_user_id,
  peer_nickname,
  peer_avatar_url,
  item_id,
  item_title,
  created_at,
  updated_at
FROM avito_chats
WHERE lead_id = $1
ORDER BY updated_at DESC
LIMIT 1
`
	chat, err := scanAvitoChat(r.db.QueryRow(ctx, query, strings.TrimSpace(leadID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AvitoChat{}, pgx.ErrNoRows
	}
	return chat, err
}

func (r *AvitoChatRepository) ListChats(ctx context.Context, userID string) ([]model.AvitoChat, error) {
	query := `
SELECT
  c.id::text,
  c.chat_id,
  c.lead_id::text,
  c.peer_user_id,
  c.peer_nickname,
  c.peer_avatar_url,
  c.item_id,
  c.item_title,
  c.created_at,
  c.updated_at,
  (
    SELECT COUNT(*)::int
    FROM avito_messages m
    WHERE m.chat_id = c.chat_id
      AND m.direction = 'incoming'
      AND m.sent_at > COALESCE(
        (
          SELECT r.last_read_at
          FROM avito_chat_reads r
          WHERE r.user_id = $1::uuid
            AND r.chat_id = c.chat_id
        ),
        TIMESTAMPTZ 'epoch'
      )
  ) AS unread_count
FROM avito_chats c
INNER JOIN leads l ON l.id = c.lead_id
  AND l.deleted_at IS NULL
  AND l.column_id <> 'failed'
ORDER BY c.updated_at DESC, c.created_at DESC
`
	rows, err := r.db.Query(ctx, query, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.AvitoChat, 0)
	for rows.Next() {
		var chat model.AvitoChat
		scanErr := rows.Scan(
			&chat.ID,
			&chat.ChatID,
			&chat.LeadID,
			&chat.PeerUserID,
			&chat.PeerNickname,
			&chat.PeerAvatarURL,
			&chat.ItemID,
			&chat.ItemTitle,
			&chat.CreatedAt,
			&chat.UpdatedAt,
			&chat.UnreadCount,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, chat)
	}
	return items, rows.Err()
}

func (r *AvitoChatRepository) InsertMessage(ctx context.Context, input model.InsertAvitoMessageInput) (model.AvitoMessage, bool, error) {
	query := `
INSERT INTO avito_messages (
  chat_id, message_id, direction, message_type, text, author_id, sent_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (chat_id, message_id) DO NOTHING
RETURNING
  id::text,
  chat_id,
  message_id,
  direction,
  message_type,
  text,
  author_id,
  sent_at,
  created_at
`
	msg, err := scanAvitoMessage(r.db.QueryRow(
		ctx,
		query,
		strings.TrimSpace(input.ChatID),
		strings.TrimSpace(input.MessageID),
		strings.TrimSpace(input.Direction),
		strings.TrimSpace(input.MessageType),
		input.Text,
		input.AuthorID,
		input.SentAt,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.AvitoMessage{}, false, nil
	}
	if err != nil {
		return model.AvitoMessage{}, false, err
	}
	return msg, true, nil
}

func (r *AvitoChatRepository) ListMessagesByChatID(ctx context.Context, chatID string) ([]model.AvitoMessage, error) {
	// Newest 200 messages — enough for UI, keeps open payload small/fast.
	query := `
SELECT
  id::text,
  chat_id,
  message_id,
  direction,
  message_type,
  text,
  author_id,
  sent_at,
  created_at
FROM (
  SELECT
    id,
    chat_id,
    message_id,
    direction,
    message_type,
    text,
    author_id,
    sent_at,
    created_at
  FROM avito_messages
  WHERE chat_id = $1
  ORDER BY sent_at DESC, created_at DESC
  LIMIT 200
) recent
ORDER BY sent_at ASC, created_at ASC
`
	rows, err := r.db.Query(ctx, query, strings.TrimSpace(chatID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.AvitoMessage, 0)
	for rows.Next() {
		msg, scanErr := scanAvitoMessage(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, msg)
	}
	return result, rows.Err()
}

func (r *AvitoChatRepository) ListMessagesByLeadID(ctx context.Context, leadID string) ([]model.AvitoMessage, error) {
	chat, err := r.GetByLeadID(ctx, leadID)
	if err != nil {
		return nil, err
	}
	return r.ListMessagesByChatID(ctx, chat.ChatID)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAvitoChat(row scannable) (model.AvitoChat, error) {
	var chat model.AvitoChat
	err := row.Scan(
		&chat.ID,
		&chat.ChatID,
		&chat.LeadID,
		&chat.PeerUserID,
		&chat.PeerNickname,
		&chat.PeerAvatarURL,
		&chat.ItemID,
		&chat.ItemTitle,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	return chat, err
}

func scanAvitoMessage(row scannable) (model.AvitoMessage, error) {
	var msg model.AvitoMessage
	err := row.Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.MessageID,
		&msg.Direction,
		&msg.MessageType,
		&msg.Text,
		&msg.AuthorID,
		&msg.SentAt,
		&msg.CreatedAt,
	)
	if err != nil {
		return model.AvitoMessage{}, err
	}
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	return msg, nil
}

func (r *AvitoChatRepository) CountUnreadChatsForUser(ctx context.Context, userID string) (int, error) {
	const query = `
SELECT COUNT(*)::int
FROM avito_chats c
INNER JOIN leads l ON l.id = c.lead_id
  AND l.deleted_at IS NULL
  AND l.column_id <> 'failed'
WHERE EXISTS (
  SELECT 1
  FROM avito_messages m
  WHERE m.chat_id = c.chat_id
    AND m.direction = 'incoming'
    AND m.sent_at > COALESCE(
      (
        SELECT r.last_read_at
        FROM avito_chat_reads r
        WHERE r.user_id = $1::uuid
          AND r.chat_id = c.chat_id
      ),
      TIMESTAMPTZ 'epoch'
    )
)
`
	var count int
	if err := r.db.QueryRow(ctx, query, strings.TrimSpace(userID)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AvitoChatRepository) MarkChatReadByLeadID(ctx context.Context, userID string, leadID string) error {
	const query = `
INSERT INTO avito_chat_reads (user_id, chat_id, last_read_at)
SELECT $1::uuid, c.chat_id, now()
FROM avito_chats c
WHERE c.lead_id = $2::uuid
ON CONFLICT (user_id, chat_id) DO UPDATE
SET last_read_at = EXCLUDED.last_read_at
`
	_, err := r.db.Exec(ctx, query, strings.TrimSpace(userID), strings.TrimSpace(leadID))
	return err
}
