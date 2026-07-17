DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'salary_service') THEN
    CREATE TYPE salary_service AS ENUM (
      'covers_without_hem',
      'covers_with_hem',
      'seat_covers',
      'steering_wheel',
      'glass_repair',
      'glass_headlight_polish'
    );
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS salary_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_date TIMESTAMPTZ NOT NULL,
  deal_id UUID NOT NULL REFERENCES deals(id),
  service salary_service NOT NULL,
  salary_amount NUMERIC(12, 2) NOT NULL CHECK (salary_amount >= 0),
  comment TEXT NOT NULL DEFAULT '',
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_salary_entries_created_by_entry_date
  ON salary_entries (created_by, entry_date DESC);

CREATE INDEX IF NOT EXISTS idx_salary_entries_deal_id
  ON salary_entries (deal_id);
