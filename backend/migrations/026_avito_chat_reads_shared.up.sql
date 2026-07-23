-- Shared read state per chat: if one employee opens a chat, it is read for everyone.

CREATE TABLE IF NOT EXISTS avito_chat_reads_shared (
  chat_id TEXT PRIMARY KEY REFERENCES avito_chats (chat_id) ON DELETE CASCADE,
  last_read_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO avito_chat_reads_shared (chat_id, last_read_at)
SELECT chat_id, MAX(last_read_at)
FROM avito_chat_reads
GROUP BY chat_id
ON CONFLICT (chat_id) DO UPDATE
SET last_read_at = GREATEST(avito_chat_reads_shared.last_read_at, EXCLUDED.last_read_at);

INSERT INTO avito_chat_reads_shared (chat_id, last_read_at)
SELECT c.chat_id, now()
FROM avito_chats c
WHERE NOT EXISTS (
  SELECT 1 FROM avito_chat_reads_shared r WHERE r.chat_id = c.chat_id
)
ON CONFLICT (chat_id) DO NOTHING;

DROP TABLE IF EXISTS avito_chat_reads;

ALTER TABLE avito_chat_reads_shared RENAME TO avito_chat_reads;
