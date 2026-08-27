-- +goose Up
CREATE TABLE generator_configs (
  id UUID PRIMARY KEY,
  actor_account_id UUID NOT NULL,
  module_name TEXT NOT NULL UNIQUE,
  source_schema TEXT NOT NULL,
  source_table TEXT NOT NULL,
  normalized_config JSONB NOT NULL,
  preview_digest TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE generator_configs;
