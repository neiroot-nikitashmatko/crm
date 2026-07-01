DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
    CREATE TYPE task_status AS ENUM ('active', 'completed');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  text TEXT NOT NULL DEFAULT '',
  due_at TIMESTAMPTZ,
  status task_status NOT NULL DEFAULT 'active',
  lead_id UUID REFERENCES leads(id),
  deal_id UUID REFERENCES deals(id),
  created_by UUID NOT NULL REFERENCES users(id),
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_tasks_owner CHECK (lead_id IS NOT NULL OR deal_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_tasks_lead_id ON tasks(lead_id);
CREATE INDEX IF NOT EXISTS idx_tasks_deal_id ON tasks(deal_id);
