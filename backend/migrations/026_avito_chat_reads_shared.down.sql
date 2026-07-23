-- Restore per-user read rows from shared chat reads.

CREATE TABLE IF NOT EXISTS avito_chat_reads_per_user (
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  chat_id TEXT NOT NULL REFERENCES avito_chats (chat_id) ON DELETE CASCADE,
  last_read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, chat_id)
);

INSERT INTO avito_chat_reads_per_user (user_id, chat_id, last_read_at)
SELECT u.id, r.chat_id, r.last_read_at
FROM users u
CROSS JOIN avito_chat_reads r
WHERE u.is_active = TRUE
ON CONFLICT (user_id, chat_id) DO NOTHING;

DROP TABLE IF EXISTS avito_chat_reads;

ALTER TABLE avito_chat_reads_per_user RENAME TO avito_chat_reads;

CREATE INDEX IF NOT EXISTS idx_avito_chat_reads_user_id ON avito_chat_reads (user_id);
