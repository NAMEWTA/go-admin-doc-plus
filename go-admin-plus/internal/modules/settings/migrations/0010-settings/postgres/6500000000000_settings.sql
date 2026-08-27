-- +goose Up
CREATE TABLE settings_values (
  id TEXT PRIMARY KEY CHECK(length(id)=36), category TEXT NOT NULL CHECK(category IN ('business','ui')),
  setting_key TEXT NOT NULL CHECK(length(setting_key) BETWEEN 3 AND 80), label TEXT NOT NULL CHECK(length(btrim(label)) BETWEEN 1 AND 120),
  label_key TEXT NOT NULL CHECK(length(label_key) BETWEEN 3 AND 1440), value TEXT NOT NULL CHECK(length(btrim(value)) BETWEEN 1 AND 500),
  description TEXT NOT NULL CHECK(length(description)<=500), enabled BOOLEAN NOT NULL, revision BIGINT NOT NULL CHECK(revision>=1),
  UNIQUE(category, setting_key)
);
CREATE INDEX settings_values_order_idx ON settings_values(category, label_key COLLATE "C", id);
CREATE TABLE settings_dictionary_types (
  id TEXT PRIMARY KEY CHECK(length(id)=36), dictionary_key TEXT NOT NULL UNIQUE CHECK(length(dictionary_key) BETWEEN 3 AND 80),
  name TEXT NOT NULL CHECK(length(btrim(name)) BETWEEN 1 AND 120), name_key TEXT NOT NULL CHECK(length(name_key) BETWEEN 3 AND 1440),
  description TEXT NOT NULL CHECK(length(description)<=500), enabled BOOLEAN NOT NULL, revision BIGINT NOT NULL CHECK(revision>=1)
);
CREATE INDEX settings_dictionary_types_order_idx ON settings_dictionary_types(name_key COLLATE "C", id);
CREATE TABLE settings_dictionary_items (
  id TEXT PRIMARY KEY CHECK(length(id)=36), dictionary_id TEXT NOT NULL REFERENCES settings_dictionary_types(id) ON DELETE RESTRICT,
  item_value TEXT NOT NULL CHECK(length(btrim(item_value)) BETWEEN 1 AND 120), value_key TEXT NOT NULL CHECK(length(value_key) BETWEEN 3 AND 1440),
  label TEXT NOT NULL CHECK(length(btrim(label)) BETWEEN 1 AND 120), label_key TEXT NOT NULL CHECK(length(label_key) BETWEEN 3 AND 1440),
  sort_order INTEGER NOT NULL CHECK(sort_order BETWEEN 0 AND 100000), enabled BOOLEAN NOT NULL, revision BIGINT NOT NULL CHECK(revision>=1),
  UNIQUE(dictionary_id, item_value)
);
CREATE INDEX settings_dictionary_items_order_idx ON settings_dictionary_items(dictionary_id, sort_order, label_key COLLATE "C", id);
