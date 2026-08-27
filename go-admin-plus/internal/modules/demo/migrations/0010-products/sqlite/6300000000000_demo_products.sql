-- +goose Up
CREATE TABLE demo_products (
  id TEXT PRIMARY KEY CHECK (length(id) = 36),
  owner_account_id TEXT NOT NULL CHECK (length(owner_account_id) BETWEEN 1 AND 64),
  sku TEXT NOT NULL UNIQUE CHECK (length(sku) BETWEEN 3 AND 32 AND sku = upper(sku) AND substr(sku, 1, 1) GLOB '[A-Z0-9]' AND sku NOT GLOB '*[^A-Z0-9_-]*'),
  name TEXT NOT NULL CHECK (length(trim(name)) BETWEEN 3 AND 120),
  name_key TEXT NOT NULL CHECK (length(name_key) BETWEEN 6 AND 960 AND name_key NOT GLOB '*[^0-9a-f]*'),
  description TEXT NOT NULL CHECK (length(description) <= 500),
  price_cents INTEGER NOT NULL CHECK (price_cents BETWEEN 0 AND 100000000),
  status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
  revision INTEGER NOT NULL DEFAULT 1 CHECK (revision >= 1),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
CREATE INDEX demo_products_owner_idx ON demo_products(owner_account_id);
CREATE INDEX demo_products_name_key_idx ON demo_products(name_key, id);
CREATE INDEX demo_products_updated_idx ON demo_products(updated_at DESC, id);
