-- +goose Up
CREATE TABLE organization_departments (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  department_key TEXT NOT NULL UNIQUE CHECK (length(department_key) BETWEEN 3 AND 64 AND department_key ~ '^[a-z0-9][a-z0-9_-]*$'),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  parent_id TEXT REFERENCES organization_departments(id) ON DELETE RESTRICT,
  sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order BETWEEN -1000000 AND 1000000),
  protected BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CHECK (parent_id IS NULL OR parent_id <> id)
);
CREATE INDEX organization_departments_parent_order_idx ON organization_departments(parent_id, sort_order, department_key);

CREATE TABLE organization_positions (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  position_key TEXT NOT NULL UNIQUE CHECK (length(position_key) BETWEEN 3 AND 64 AND position_key ~ '^[a-z0-9][a-z0-9_-]*$'),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  department_id TEXT NOT NULL REFERENCES organization_departments(id) ON DELETE RESTRICT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  protected BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX organization_positions_department_idx ON organization_positions(department_id, position_key);

INSERT INTO organization_departments(id, department_key, name, parent_id, sort_order, protected, created_at, updated_at)
VALUES ('department-root-001', 'root', 'Organization', NULL, 0, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
