CREATE TABLE IF NOT EXISTS avito_chats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id TEXT NOT NULL UNIQUE,
  lead_id UUID NOT NULL REFERENCES leads(id),
  peer_user_id BIGINT,
  peer_nickname TEXT NOT NULL DEFAULT '',
  peer_avatar_url TEXT NOT NULL DEFAULT '',
  item_id BIGINT,
  item_title TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_avito_chats_lead_id ON avito_chats (lead_id);

CREATE TABLE IF NOT EXISTS avito_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id TEXT NOT NULL REFERENCES avito_chats(chat_id) ON DELETE CASCADE,
  message_id TEXT NOT NULL,
  direction TEXT NOT NULL CHECK (direction IN ('incoming', 'outgoing')),
  message_type TEXT NOT NULL DEFAULT 'text',
  text TEXT NOT NULL DEFAULT '',
  author_id BIGINT,
  sent_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (chat_id, message_id)
);

CREATE INDEX IF NOT EXISTS idx_avito_messages_chat_id_sent_at
  ON avito_messages (chat_id, sent_at ASC);
