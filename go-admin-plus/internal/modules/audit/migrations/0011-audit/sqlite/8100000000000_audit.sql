-- +goose Up
CREATE TABLE audit_facts (
    topic TEXT NOT NULL CHECK (topic IN (
        'iam.login.succeeded',
        'iam.login.failed',
		'operation.created',
		'operation.updated',
		'operation.deleted'
    )),
    business_key TEXT NOT NULL CHECK (length(business_key) BETWEEN 3 AND 255) CHECK (
		(topic IN ('iam.login.succeeded', 'iam.login.failed')
			AND substr(business_key, 1, 6) = 'login:'
			AND length(substr(business_key, 7)) = 32
			AND substr(business_key, 7) NOT GLOB '*[^a-f0-9]*')
		OR
		(topic IN ('operation.created', 'operation.updated', 'operation.deleted') AND (
			(length(business_key) - length(replace(business_key, ':', '')) = 4
				AND substr(business_key, -7) = ':system')
			OR
			(length(business_key) - length(replace(business_key, ':', '')) = 5
				AND instr(business_key, ':account:') > 0
				AND length(substr(business_key, instr(business_key, ':account:') + 9)) BETWEEN 8 AND 64
				AND substr(business_key, instr(business_key, ':account:') + 9) NOT GLOB '*[^a-z0-9_-]*')
		))
	),
    actor_ref TEXT CHECK (
        length(actor_ref) BETWEEN 16 AND 72
        AND substr(actor_ref, 1, 8) = 'account:'
        AND substr(actor_ref, 9, 1) GLOB '[a-z0-9]'
        AND substr(actor_ref, 9) NOT GLOB '*[^a-z0-9_-]*'
    ),
    payload BLOB NOT NULL CHECK (length(payload) BETWEEN 2 AND 1024),
	occurred_at TIMESTAMP NOT NULL,
	PRIMARY KEY (topic, business_key)
);

CREATE INDEX audit_facts_occurred_idx
	ON audit_facts (occurred_at DESC, topic, business_key);
CREATE INDEX audit_facts_topic_occurred_idx
	ON audit_facts (topic, occurred_at DESC, business_key);
