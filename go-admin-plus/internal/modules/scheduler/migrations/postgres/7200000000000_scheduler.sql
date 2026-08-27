-- +goose Up
CREATE TABLE scheduler_definitions (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    name_key text COLLATE "C" NOT NULL UNIQUE,
    task_type text COLLATE "C" NOT NULL,
    schedule_json bytea NOT NULL,
    parameters_json bytea NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision >= 1),
    next_run_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz,
    CHECK ((enabled AND next_run_at IS NOT NULL AND deleted_at IS NULL) OR (NOT enabled AND next_run_at IS NULL))
);
CREATE INDEX scheduler_definitions_due_idx ON scheduler_definitions(enabled, next_run_at, id) WHERE deleted_at IS NULL;
CREATE TABLE scheduler_executions (
    id uuid PRIMARY KEY,
    definition_id uuid NOT NULL REFERENCES scheduler_definitions(id),
    definition_revision bigint NOT NULL CHECK (definition_revision >= 1),
    task_type text COLLATE "C" NOT NULL,
    scheduled_for timestamptz NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    status text COLLATE "C" NOT NULL CHECK (status IN ('succeeded', 'failed')),
    error_code text COLLATE "C",
    executor_owner text COLLATE "C" NOT NULL,
    CHECK (scheduled_for <= started_at),
    CHECK (started_at <= finished_at),
    CHECK ((status = 'succeeded' AND error_code IS NULL) OR (status = 'failed' AND error_code IS NOT NULL)),
    UNIQUE(definition_id, scheduled_for)
);
CREATE INDEX scheduler_executions_history_idx ON scheduler_executions(started_at DESC, id DESC);
CREATE INDEX scheduler_executions_definition_idx ON scheduler_executions(definition_id, started_at DESC, id DESC);
