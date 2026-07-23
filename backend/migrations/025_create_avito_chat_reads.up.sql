CREATE TABLE IF NOT EXISTS avito_chat_reads (
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  chat_id TEXT NOT NULL REFERENCES avito_chats (chat_id) ON DELETE CASCADE,
  last_read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, chat_id)
);

CREATE INDEX IF NOT EXISTS idx_avito_chat_reads_user_id ON avito_chat_reads (user_id);

-- Existing chats start as read, so deploy does not flood badges with old history.
INSERT INTO avito_chat_reads (user_id, chat_id, last_read_at)
SELECT u.id, c.chat_id, now()
FROM users u
CROSS JOIN avito_chats c
WHERE u.is_active = TRUE
ON CONFLICT (user_id, chat_id) DO NOTHING;
