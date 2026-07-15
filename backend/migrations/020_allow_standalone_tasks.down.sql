DELETE FROM tasks WHERE lead_id IS NULL AND deal_id IS NULL;

ALTER TABLE tasks
  ADD CONSTRAINT chk_tasks_owner CHECK (lead_id IS NOT NULL OR deal_id IS NOT NULL);
