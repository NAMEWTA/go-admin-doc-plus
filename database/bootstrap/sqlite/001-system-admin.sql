PRAGMA foreign_keys = ON;

BEGIN IMMEDIATE;

-- One-time bootstrap credential: admin / administrator password
-- Re-running this script grants the system-admin role but never resets an
-- existing account password.
INSERT INTO iam_accounts (
  id,
  username,
  display_name,
  email,
  password_hash,
  password_changed_at,
  created_at,
  updated_at
) VALUES (
  'account-system-admin',
  'admin',
  'Administrator',
  'admin@example.test',
  '$argon2id$v=19$m=65536,t=3,p=4$jrA0Z/FfoK10mrfgLeWzrQ$319NgrCZqJlZVVkfnftRAaKQW4HYrAKxEJ2vNLtiCT0',
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
) ON CONFLICT(username) DO NOTHING;

INSERT INTO iam_account_roles (account_id, role_id)
SELECT id, 'role-system-admin'
FROM iam_accounts
WHERE username = 'admin'
ON CONFLICT(account_id, role_id) DO NOTHING;

COMMIT;
