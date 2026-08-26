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
    actor_ref TEXT CHECK (
        length(actor_ref) BETWEEN 9 AND 72
        AND substr(actor_ref, 1, 8) = 'account:'
        AND substr(actor_ref, 9, 1) GLOB '[a-z0-9]'
        AND substr(actor_ref, 9) NOT GLOB '*[^a-z0-9_-]*'
    ),
    payload BLOB NOT NULL CHECK (length(payload) BETWEEN 2 AND 1024),
    occurred_at TIMESTAMP NOT NULL
);

CREATE INDEX audit_facts_occurred_idx
    ON audit_facts (occurred_at DESC, event_id DESC);
CREATE INDEX audit_facts_topic_occurred_idx
    ON audit_facts (topic, occurred_at DESC, event_id DESC);
