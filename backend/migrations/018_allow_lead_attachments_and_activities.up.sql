ALTER TABLE attachments
  DROP CONSTRAINT IF EXISTS attachments_entity_type_check;

ALTER TABLE attachments
  ADD CONSTRAINT attachments_entity_type_check
  CHECK (entity_type IN ('deal', 'task', 'lead'));

ALTER TABLE activities
  DROP CONSTRAINT IF EXISTS activities_entity_type_check;

ALTER TABLE activities
  ADD CONSTRAINT activities_entity_type_check
  CHECK (entity_type IN ('deal', 'task', 'lead'));
