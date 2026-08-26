PRAGMA foreign_keys = ON;

CREATE TABLE sys_migration (
  version TEXT,
  apply_time DATETIME,
  PRIMARY KEY (version)
);

CREATE TABLE demo_product (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  code TEXT,
  price REAL,
  status TEXT,
  remark TEXT,
  create_by INTEGER,
  update_by INTEGER,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_demo_product_update_by ON demo_product (update_by);
CREATE INDEX idx_demo_product_create_by ON demo_product (create_by);

INSERT INTO sys_migration (version, apply_time)
VALUES ('1786700000000', '2026-08-24T00:00:00Z');

INSERT INTO demo_product (
  id,
  name,
  code,
  price,
  status,
  remark,
  create_by,
  update_by,
  created_at,
  updated_at,
  deleted_at
) VALUES (
  1,
  'Pre Phase One Product',
  'PRE-PHASE1-001',
  19.95,
  '2',
  'Deterministic migration fixture',
  1,
  1,
  '2026-08-24T00:00:00Z',
  '2026-08-24T00:00:00Z',
  0
);
