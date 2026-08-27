-- +goose Up
CREATE TABLE scheduler_definitions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL UNIQUE,
    task_type TEXT NOT NULL,
    schedule_json BLOB NOT NULL,
    parameters_json BLOB NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0, 1)),
    revision INTEGER NOT NULL DEFAULT 1 CHECK (revision >= 1),
    next_run_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    CHECK ((enabled = 1 AND next_run_at IS NOT NULL AND deleted_at IS NULL) OR (enabled = 0 AND next_run_at IS NULL))
);
CREATE INDEX scheduler_definitions_due_idx ON scheduler_definitions(enabled, next_run_at, id) WHERE deleted_at IS NULL;
CREATE TABLE scheduler_executions (
    id TEXT PRIMARY KEY,
    definition_id TEXT NOT NULL REFERENCES scheduler_definitions(id),
    definition_revision INTEGER NOT NULL CHECK (definition_revision >= 1),
    task_type TEXT NOT NULL,
    scheduled_for TIMESTAMP NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('succeeded', 'failed')),
    error_code TEXT,
    executor_owner TEXT NOT NULL,
    CHECK (scheduled_for <= started_at),
    CHECK (started_at <= finished_at),
    CHECK ((status = 'succeeded' AND error_code IS NULL) OR (status = 'failed' AND error_code IS NOT NULL)),
    UNIQUE(definition_id, scheduled_for)
);
CREATE INDEX scheduler_executions_history_idx ON scheduler_executions(started_at DESC, id DESC);
CREATE INDEX scheduler_executions_definition_idx ON scheduler_executions(definition_id, started_at DESC, id DESC);
