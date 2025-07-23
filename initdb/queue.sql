CREATE TYPE task_status AS ENUM ('CREATED', 'READY', 'PROCESSING', 'DELAYED', 'COMPLETED', 'FAILED');

CREATE TABLE tasks (
    id uuid PRIMARY KEY,
    kind varchar(255) NOT NULL,
    created_at timestamptz NOT NULL,
    finalized_at timestamptz NULL,
    status task_status NOT NULL,
    status_changed_at timestamptz NOT NULL,
    delayed_until timestamptz NULL,
    timeout_at timestamptz NULL,
    priority smallint NOT NULL,
    retries int NOT NULL,
    version int NOT NULL
);

CREATE INDEX ON tasks (kind, status, priority DESC, status_changed_at ASC) WHERE status = 'READY';
CREATE INDEX ON tasks (status, delayed_until) WHERE status = 'DELAYED';
CREATE INDEX ON tasks (status, timeout_at) WHERE status = 'PROCESSING';
CREATE INDEX ON tasks (status, finalized_at) WHERE status IN ('COMPLETED', 'FAILED');
CREATE INDEX ON tasks (created_at);

CREATE TABLE task_payloads (
    task_id uuid PRIMARY KEY,
    payload jsonb NOT NULL
);

CREATE TABLE task_results (
    task_id uuid PRIMARY KEY,
    result jsonb NOT NULL
);

CREATE TABLE archived_tasks (
    id uuid PRIMARY KEY,
    kind varchar(255) NOT NULL,
    created_at timestamptz NOT NULL,
    finalized_at timestamptz NOT NULL,
    status task_status NOT NULL,
    priority smallint NOT NULL,
    retries int NOT NULL,
    payload jsonb NOT NULL,
    result jsonb NULL
);
