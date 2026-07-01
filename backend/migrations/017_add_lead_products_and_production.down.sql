DROP TABLE IF EXISTS lead_products;

ALTER TABLE leads
  DROP COLUMN IF EXISTS production_employee,
  DROP COLUMN IF EXISTS production_due_at,
  DROP COLUMN IF EXISTS production_nomenclature;
