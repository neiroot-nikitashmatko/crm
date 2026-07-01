DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'deal_status') THEN
    CREATE TYPE deal_status AS ENUM ('today', 'tomorrow', 'later', 'closed', 'failed');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS deals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  deal_number BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
  lead_id UUID REFERENCES leads(id),
  first_name TEXT NOT NULL,
  patronymic TEXT,
  phone TEXT NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{10,14}$'),
  traffic_source TEXT NOT NULL DEFAULT '',
  status deal_status NOT NULL DEFAULT 'today',
  total_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  deal_comments TEXT NOT NULL DEFAULT '',
  production_nomenclature TEXT NOT NULL DEFAULT '',
  production_due_at TIMESTAMPTZ,
  production_employee TEXT NOT NULL DEFAULT '',
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_deals_active_per_lead
  ON deals (lead_id)
  WHERE lead_id IS NOT NULL AND status <> 'failed';
