INSERT INTO activities (entity_type, entity_id, author_id, activity_type, text, created_at)
SELECT
  'lead',
  l.id,
  l.created_by,
  'system',
  'Лид создан',
  l.created_at
FROM leads l
WHERE l.deleted_at IS NULL
  AND l.created_by IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM activities a
    WHERE a.entity_type = 'lead'
      AND a.entity_id = l.id
      AND a.activity_type = 'system'
      AND a.text = 'Лид создан'
      AND a.deleted_at IS NULL
  );
