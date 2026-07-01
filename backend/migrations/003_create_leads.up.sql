CREATE TABLE IF NOT EXISTS leads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  lead_number BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
  first_name TEXT NOT NULL,
  patronymic TEXT,
  phone TEXT NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{10,14}$'),
  traffic_source TEXT NOT NULL DEFAULT '',
  column_id TEXT NOT NULL CHECK (column_id IN ('new', 'chat', 'phone', 'deal', 'failed')),
  lead_comments TEXT NOT NULL DEFAULT '',
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_leads_column_id ON leads(column_id);
