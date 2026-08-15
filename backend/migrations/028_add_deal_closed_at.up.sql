ALTER TABLE deals
  ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ;

UPDATE deals
SET closed_at = updated_at
WHERE status = 'closed'
  AND closed_at IS NULL;

UPDATE deals
SET closed_at = NULL
WHERE status <> 'closed'
  AND closed_at IS NOT NULL;

ALTER TABLE deals
  DROP CONSTRAINT IF EXISTS deals_closed_at_status_check;

ALTER TABLE deals
  ADD CONSTRAINT deals_closed_at_status_check
  CHECK (
    (status = 'closed' AND closed_at IS NOT NULL)
    OR (status <> 'closed' AND closed_at IS NULL)
  );

CREATE INDEX IF NOT EXISTS idx_deals_closed_at
  ON deals (closed_at DESC)
  WHERE deleted_at IS NULL AND status = 'closed';
