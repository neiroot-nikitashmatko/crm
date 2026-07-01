CREATE TABLE IF NOT EXISTS catalog_products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  sku TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  cost NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (cost >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_catalog_products_sku ON catalog_products (sku);
CREATE INDEX IF NOT EXISTS idx_catalog_products_category ON catalog_products (category);
CREATE INDEX IF NOT EXISTS idx_catalog_products_name ON catalog_products (name);
