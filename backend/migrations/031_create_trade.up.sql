CREATE TABLE IF NOT EXISTS suppliers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  contact_person TEXT NOT NULL,
  phone TEXT NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{10,14}$'),
  inn TEXT NOT NULL,
  kpp TEXT NOT NULL DEFAULT '',
  ogrn TEXT NOT NULL DEFAULT '',
  legal_address TEXT NOT NULL,
  actual_address TEXT NOT NULL,
  bik TEXT NOT NULL,
  settlement_account TEXT NOT NULL,
  correspondent_account TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers (name);

CREATE TABLE IF NOT EXISTS incoming_invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_number BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
  invoice_date TIMESTAMPTZ NOT NULL,
  supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
  total NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (total >= 0),
  comment TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_incoming_invoices_supplier_id ON incoming_invoices (supplier_id);
CREATE INDEX IF NOT EXISTS idx_incoming_invoices_date ON incoming_invoices (invoice_date DESC);

CREATE TABLE IF NOT EXISTS incoming_invoice_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES incoming_invoices(id) ON DELETE CASCADE,
  position INTEGER NOT NULL DEFAULT 0,
  catalog_product_id UUID REFERENCES catalog_products(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  quantity NUMERIC(12, 3) NOT NULL CHECK (quantity > 0),
  unit_price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_incoming_invoice_items_position
  ON incoming_invoice_items (invoice_id, position);

CREATE INDEX IF NOT EXISTS idx_incoming_invoice_items_catalog_product_id
  ON incoming_invoice_items (catalog_product_id);

CREATE TABLE IF NOT EXISTS outgoing_invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_number BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
  invoice_date TIMESTAMPTZ NOT NULL,
  deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE RESTRICT,
  total NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (total >= 0),
  comment TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_outgoing_invoices_deal_id UNIQUE (deal_id)
);

CREATE INDEX IF NOT EXISTS idx_outgoing_invoices_date ON outgoing_invoices (invoice_date DESC);

CREATE TABLE IF NOT EXISTS outgoing_invoice_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id UUID NOT NULL REFERENCES outgoing_invoices(id) ON DELETE CASCADE,
  position INTEGER NOT NULL DEFAULT 0,
  catalog_product_id UUID REFERENCES catalog_products(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  quantity NUMERIC(12, 3) NOT NULL CHECK (quantity > 0),
  unit_price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_outgoing_invoice_items_position
  ON outgoing_invoice_items (invoice_id, position);

CREATE INDEX IF NOT EXISTS idx_outgoing_invoice_items_catalog_product_id
  ON outgoing_invoice_items (catalog_product_id);
