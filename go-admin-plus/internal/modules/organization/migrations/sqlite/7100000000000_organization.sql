-- +goose Up
CREATE TABLE organization_departments (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  department_key TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(department_key) BETWEEN 3 AND 64 AND department_key NOT GLOB '*[^a-z0-9_-]*' AND substr(department_key, 1, 1) GLOB '[a-z0-9]'),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  parent_id TEXT REFERENCES organization_departments(id) ON DELETE RESTRICT,
  sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order BETWEEN -1000000 AND 1000000),
  protected INTEGER NOT NULL DEFAULT 0 CHECK (protected IN (0, 1)),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE INDEX organization_departments_parent_order_idx ON organization_departments(parent_id, sort_order, department_key);

CREATE TABLE organization_positions (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  position_key TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(position_key) BETWEEN 3 AND 64 AND position_key NOT GLOB '*[^a-z0-9_-]*' AND substr(position_key, 1, 1) GLOB '[a-z0-9]'),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  name_key TEXT NOT NULL CHECK (length(name_key) BETWEEN 3 AND 1200 AND length(name_key) % 3 = 0 AND substr(name_key, -1, 1) = '.' AND name_key NOT GLOB '*[^0-9a-f.]*'),
  department_id TEXT NOT NULL REFERENCES organization_departments(id) ON DELETE RESTRICT,
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  protected INTEGER NOT NULL DEFAULT 0 CHECK (protected IN (0, 1)),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
CREATE INDEX organization_positions_department_idx ON organization_positions(department_id, position_key);
CREATE INDEX organization_positions_name_key_idx ON organization_positions(name_key, position_key);

INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at)
VALUES ('department-root-001', 'root', 'Organization', NULL, 0, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
