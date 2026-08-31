-- +goose Up
CREATE TABLE iam_role_data_scopes (
  role_id TEXT PRIMARY KEY REFERENCES iam_roles(id) ON DELETE CASCADE,
  scope TEXT NOT NULL CHECK (scope IN ('all','self','organization','organization-tree','custom')),
  updated_at TIMESTAMP NOT NULL
);
INSERT INTO iam_role_data_scopes(role_id, scope, updated_at)
SELECT id, data_scope, CURRENT_TIMESTAMP FROM iam_roles;

CREATE TABLE iam_account_organization (
  account_id TEXT PRIMARY KEY REFERENCES iam_accounts(id) ON DELETE CASCADE,
  primary_department_id TEXT REFERENCES organization_departments(id) ON DELETE RESTRICT,
  updated_at TIMESTAMP NOT NULL
);
CREATE INDEX iam_account_organization_department_idx ON iam_account_organization(primary_department_id);

CREATE TABLE iam_account_positions (
  account_id TEXT NOT NULL REFERENCES iam_accounts(id) ON DELETE CASCADE,
  position_id TEXT NOT NULL REFERENCES organization_positions(id) ON DELETE RESTRICT,
  PRIMARY KEY (account_id, position_id)
);
CREATE INDEX iam_account_positions_position_idx ON iam_account_positions(position_id);

CREATE TABLE iam_role_data_scope_departments (
  role_id TEXT NOT NULL REFERENCES iam_role_data_scopes(role_id) ON DELETE CASCADE,
  department_id TEXT NOT NULL REFERENCES organization_departments(id) ON DELETE RESTRICT,
  PRIMARY KEY (role_id, department_id)
);
CREATE INDEX iam_role_data_scope_departments_department_idx ON iam_role_data_scope_departments(department_id);
