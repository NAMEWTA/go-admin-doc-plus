CREATE TABLE goose_db_version (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL,
    is_applied INTEGER NOT NULL,
    tstamp TIMESTAMP DEFAULT (datetime('now'))
);
INSERT INTO goose_db_version (version_id, is_applied) VALUES (0, 1);
INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1);
CREATE TABLE architecture_marker (
    id INTEGER PRIMARY KEY,
    generation TEXT NOT NULL
);
INSERT INTO architecture_marker (id, generation) VALUES (1, 'previous-greenfield');
