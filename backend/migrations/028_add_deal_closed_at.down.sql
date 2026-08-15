DROP INDEX IF EXISTS idx_deals_closed_at;

ALTER TABLE deals
  DROP CONSTRAINT IF EXISTS deals_closed_at_status_check;

ALTER TABLE deals
  DROP COLUMN IF EXISTS closed_at;
