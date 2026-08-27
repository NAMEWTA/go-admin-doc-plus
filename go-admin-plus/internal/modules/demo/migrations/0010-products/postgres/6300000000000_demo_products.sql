-- +goose Up
CREATE TABLE demo_products (
  id TEXT PRIMARY KEY CHECK (length(id) = 36),
  owner_account_id TEXT NOT NULL CHECK (length(owner_account_id) BETWEEN 1 AND 64),
  sku TEXT NOT NULL UNIQUE CHECK (sku ~ '^[A-Z0-9][A-Z0-9_-]{2,31}$'),
  name TEXT NOT NULL CHECK (length(btrim(name)) BETWEEN 3 AND 120),
  name_key TEXT NOT NULL CHECK (length(name_key) BETWEEN 9 AND 1440 AND name_key ~ '^([0-9a-f]{2}\.)+$'),
  description TEXT NOT NULL CHECK (length(description) <= 500),
  price_cents BIGINT NOT NULL CHECK (price_cents BETWEEN 0 AND 100000000),
  status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
  revision BIGINT NOT NULL DEFAULT 1 CHECK (revision >= 1),
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX demo_products_owner_idx ON demo_products(owner_account_id);
CREATE INDEX demo_products_name_key_idx ON demo_products(name_key, id);
CREATE INDEX demo_products_updated_idx ON demo_products(updated_at DESC, id);
