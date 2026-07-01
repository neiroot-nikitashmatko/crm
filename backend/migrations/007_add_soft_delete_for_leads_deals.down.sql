DROP INDEX IF EXISTS uq_deals_active_per_lead;
CREATE UNIQUE INDEX uq_deals_active_per_lead
  ON deals (lead_id)
  WHERE lead_id IS NOT NULL AND status <> 'failed';

ALTER TABLE leads DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE deals DROP COLUMN IF EXISTS deleted_at;
