ALTER TABLE leads
  DROP COLUMN IF EXISTS pickup_address,
  DROP COLUMN IF EXISTS pickup_date,
  DROP COLUMN IF EXISTS delivery_address,
  DROP COLUMN IF EXISTS delivery_date,
  DROP COLUMN IF EXISTS courier;
