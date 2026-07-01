ALTER TABLE leads
  ADD COLUMN IF NOT EXISTS production_nomenclature TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS production_due_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS production_employee TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS lead_products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
  position INTEGER NOT NULL DEFAULT 0,
  title TEXT NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 1,
  unit_price NUMERIC(12,2) NOT NULL DEFAULT 0,
  amount NUMERIC(12,2) NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_lead_products_position
  ON lead_products(lead_id, position);

CREATE INDEX IF NOT EXISTS idx_lead_products_lead_id ON lead_products(lead_id);
