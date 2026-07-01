CREATE TABLE IF NOT EXISTS attachments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type TEXT NOT NULL CHECK (entity_type IN ('deal', 'task')),
  entity_id UUID NOT NULL,
  name TEXT NOT NULL,
  size BIGINT NOT NULL CHECK (size > 0),
  mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
  content BYTEA NOT NULL,
  uploaded_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_attachments_entity
  ON attachments (entity_type, entity_id)
  WHERE deleted_at IS NULL;
