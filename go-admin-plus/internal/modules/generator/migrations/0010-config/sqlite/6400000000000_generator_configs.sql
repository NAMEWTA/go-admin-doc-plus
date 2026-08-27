-- +goose Up
CREATE TABLE generator_configs (
  id TEXT PRIMARY KEY,
  actor_account_id TEXT NOT NULL,
  module_name TEXT NOT NULL UNIQUE,
  source_schema TEXT NOT NULL,
  source_table TEXT NOT NULL,
  normalized_config TEXT NOT NULL,
  preview_digest TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- +goose Down
DROP TABLE generator_configs;
