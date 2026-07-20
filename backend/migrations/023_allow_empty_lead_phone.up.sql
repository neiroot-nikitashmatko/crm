-- Allow empty phone for leads without a real number (e.g. Avito chat).
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_phone_check;
ALTER TABLE leads
  ADD CONSTRAINT leads_phone_check
  CHECK (phone = '' OR phone ~ '^\+[1-9][0-9]{10,14}$');
