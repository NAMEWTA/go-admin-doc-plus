-- +goose Up
CREATE TABLE iam_permissions (
  code TEXT PRIMARY KEY CHECK (length(code) BETWEEN 3 AND 100),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  protected INTEGER NOT NULL DEFAULT 1 CHECK (protected IN (0, 1))
);
CREATE TABLE iam_roles (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  role_key TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(role_key) BETWEEN 3 AND 64 AND role_key NOT GLOB '*[^a-z0-9_-]*' AND substr(role_key, 1, 1) GLOB '[a-z0-9]'),
  name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 100),
  data_scope TEXT NOT NULL CHECK (data_scope IN ('all', 'self')),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
  protected INTEGER NOT NULL DEFAULT 0 CHECK (protected IN (0, 1)),
  created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL
);
CREATE TABLE iam_account_roles (
  account_id TEXT NOT NULL REFERENCES iam_accounts(id) ON DELETE CASCADE,
  role_id TEXT NOT NULL REFERENCES iam_roles(id) ON DELETE RESTRICT,
  PRIMARY KEY (account_id, role_id)
);
CREATE TABLE iam_role_permissions (
  role_id TEXT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  permission_code TEXT NOT NULL REFERENCES iam_permissions(code) ON DELETE RESTRICT,
  PRIMARY KEY (role_id, permission_code)
);
CREATE TABLE iam_menus (
  id TEXT PRIMARY KEY CHECK (length(id) BETWEEN 16 AND 64),
  menu_key TEXT NOT NULL COLLATE NOCASE UNIQUE CHECK (length(menu_key) BETWEEN 3 AND 64 AND menu_key NOT GLOB '*[^a-z0-9_-]*' AND substr(menu_key, 1, 1) GLOB '[a-z0-9]'),
  label TEXT NOT NULL CHECK (length(label) BETWEEN 1 AND 80),
  path TEXT NOT NULL UNIQUE CHECK (substr(path, 1, 1) = '/'),
  permission_code TEXT NOT NULL REFERENCES iam_permissions(code) ON DELETE RESTRICT,
  sort_order INTEGER NOT NULL DEFAULT 0,
  protected INTEGER NOT NULL DEFAULT 0 CHECK (protected IN (0, 1)),
  created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL
);
CREATE TABLE iam_role_menus (
  role_id TEXT NOT NULL REFERENCES iam_roles(id) ON DELETE CASCADE,
  menu_id TEXT NOT NULL REFERENCES iam_menus(id) ON DELETE RESTRICT,
  PRIMARY KEY (role_id, menu_id)
);
CREATE INDEX iam_account_roles_role_idx ON iam_account_roles(role_id);
CREATE INDEX iam_role_permissions_permission_idx ON iam_role_permissions(permission_code);
CREATE INDEX iam_role_menus_menu_idx ON iam_role_menus(menu_id);

INSERT INTO iam_permissions(code, name) VALUES
 ('iam.users.read', 'Read users'), ('iam.users.write', 'Manage users'),
 ('iam.users.delete', 'Delete users'), ('iam.users.reset-password', 'Reset user passwords'),
 ('iam.roles.read', 'Read roles'), ('iam.roles.write', 'Manage roles'),
 ('iam.roles.delete', 'Delete roles'), ('iam.roles.assign', 'Assign authorization'),
 ('iam.menus.read', 'Read menus'), ('iam.menus.write', 'Manage menus'),
 ('iam.menus.delete', 'Delete menus'), ('iam.permissions.read', 'Read permission codes'),
 ('iam.manifest.read', 'Read capability manifest');
INSERT INTO iam_roles(id, role_key, name, data_scope, enabled, protected, created_at, updated_at)
 VALUES ('role-system-admin', 'system-admin', 'System administrator', 'all', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
INSERT INTO iam_role_permissions(role_id, permission_code)
 SELECT 'role-system-admin', code FROM iam_permissions;
INSERT INTO iam_menus(id, menu_key, label, path, permission_code, sort_order, protected, created_at, updated_at) VALUES
 ('menu-iam-users-01', 'iam-users', 'Users', '/iam/users', 'iam.users.read', 10, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
 ('menu-iam-roles-01', 'iam-roles', 'Roles', '/iam/roles', 'iam.roles.read', 20, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
 ('menu-iam-menus-01', 'iam-menus', 'Menus', '/iam/menus', 'iam.menus.read', 30, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
INSERT INTO iam_role_menus(role_id, menu_id) SELECT 'role-system-admin', id FROM iam_menus;
