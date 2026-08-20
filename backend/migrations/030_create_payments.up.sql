DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_payer') THEN
    CREATE TYPE payment_payer AS ENUM (
      'ip-panov-nikolay',
      'ip-shmatko-nikita',
      'ip-panov-dmitry'
    );
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  short_title TEXT NOT NULL
    CHECK (btrim(short_title) <> '' AND char_length(short_title) <= 22),
  payment_date TIMESTAMPTZ NOT NULL,
  remind_at TIMESTAMPTZ,
  payer_id payment_payer,
  counterparty TEXT NOT NULL DEFAULT '',
  amount NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
  comment TEXT NOT NULL DEFAULT '',
  is_closed BOOLEAN NOT NULL DEFAULT FALSE,
  closed_at TIMESTAMPTZ,
  reminder_sent_at TIMESTAMPTZ,
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT payments_closed_at_consistent CHECK (
    (is_closed = FALSE AND closed_at IS NULL)
    OR (is_closed = TRUE AND closed_at IS NOT NULL)
  ),
  CONSTRAINT payments_remind_not_after_payment CHECK (
    remind_at IS NULL
    OR (timezone('Europe/Moscow', remind_at))::date
       <= (timezone('Europe/Moscow', payment_date))::date
  ),
  CONSTRAINT payments_reminder_sent_requires_remind CHECK (
    reminder_sent_at IS NULL OR remind_at IS NOT NULL
  )
);

CREATE INDEX IF NOT EXISTS idx_payments_payment_date
  ON payments (payment_date);

CREATE INDEX IF NOT EXISTS idx_payments_created_by
  ON payments (created_by);

CREATE INDEX IF NOT EXISTS idx_payments_due_reminders
  ON payments (remind_at)
  WHERE remind_at IS NOT NULL
    AND reminder_sent_at IS NULL
    AND is_closed = FALSE;
