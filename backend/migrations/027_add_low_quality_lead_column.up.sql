ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_column_id_check;

ALTER TABLE leads
ADD CONSTRAINT leads_column_id_check
CHECK (column_id IN ('new', 'chat', 'phone', 'deal', 'failed', 'low_quality'));
