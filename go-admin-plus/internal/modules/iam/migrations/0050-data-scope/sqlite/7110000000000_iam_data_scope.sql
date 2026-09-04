-- +goose Up
CREATE TABLE iam_role_data_scopes (
  role_id TEXT PRIMARY KEY REFERENCES iam_roles(id) ON DELETE CASCADE,
  scope TEXT NOT NULL CHECK (scope IN ('all','self')),
  updated_at TIMESTAMP NOT NULL
);
INSERT INTO iam_role_data_scopes(role_id, scope, updated_at)
SELECT id, data_scope, CURRENT_TIMESTAMP FROM iam_roles;
