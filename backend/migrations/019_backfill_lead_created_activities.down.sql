DELETE FROM activities
WHERE entity_type = 'lead'
  AND activity_type = 'system'
  AND text = 'Лид создан';
