ALTER TABLE leads ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE deals ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

DROP INDEX IF EXISTS uq_deals_active_per_lead;
CREATE UNIQUE INDEX uq_deals_active_per_lead
  ON deals (lead_id)
  WHERE lead_id IS NOT NULL AND status <> 'failed' AND deleted_at IS NULL;
