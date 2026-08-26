-- +goose Up
CREATE TABLE audit_facts (
    event_id TEXT PRIMARY KEY CHECK (length(event_id) BETWEEN 16 AND 64),
    topic TEXT NOT NULL CHECK (topic IN (
        'iam.login.succeeded',
        'iam.login.failed',
		'operation.created',
		'operation.updated',
		'operation.deleted'
    )),
    business_key TEXT NOT NULL CHECK (length(business_key) BETWEEN 3 AND 255),
    actor_ref TEXT CHECK (actor_ref ~ '^account:[a-z0-9][a-z0-9_-]{0,63}$'),
    payload BYTEA NOT NULL CHECK (octet_length(payload) BETWEEN 2 AND 1024),
    occurred_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX audit_facts_occurred_idx
    ON audit_facts (occurred_at DESC, event_id DESC);
CREATE INDEX audit_facts_topic_occurred_idx
    ON audit_facts (topic, occurred_at DESC, event_id DESC);
