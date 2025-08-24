CREATE TYPE message_status AS ENUM ('CREATED', 'READY', 'PROCESSING', 'DELAYED', 'COMPLETED', 'FAILED');

CREATE TABLE messages (
    id uuid PRIMARY KEY,
    queue varchar(255) NOT NULL,
    created_at timestamptz NOT NULL,
    finalized_at timestamptz NULL,
    status message_status NOT NULL,
    status_changed_at timestamptz NOT NULL,
    delayed_until timestamptz NULL,
    timeout_at timestamptz NULL,
    priority smallint NOT NULL,
    retries int NOT NULL,
    version int NOT NULL
);

CREATE INDEX ON messages (queue, status, priority DESC, status_changed_at ASC) WHERE status = 'READY';
CREATE INDEX ON messages (status, delayed_until) WHERE status = 'DELAYED';
CREATE INDEX ON messages (status, timeout_at) WHERE status = 'PROCESSING';
CREATE INDEX ON messages (status, finalized_at) WHERE status IN ('COMPLETED', 'FAILED');
CREATE INDEX ON messages (created_at);

CREATE TABLE message_payloads (
    msg_id uuid PRIMARY KEY,
    payload text NOT NULL
);

CREATE TABLE archived_messages (
    id uuid PRIMARY KEY,
    queue varchar(255) NOT NULL,
    created_at timestamptz NOT NULL,
    finalized_at timestamptz NOT NULL,
    status message_status NOT NULL,
    priority smallint NOT NULL,
    retries int NOT NULL,
    payload text NOT NULL
);
