-- OpenClaw Oversight Agent schema
-- Tables: actions (shared, owned by Team 1), boundaries, violations

-- Shared actions table (created by Team 1, included here for reference)
CREATE TABLE IF NOT EXISTS actions (
    id            SERIAL PRIMARY KEY,
    agent_id      TEXT        NOT NULL,
    action_type   TEXT        NOT NULL,
    payload       JSONB       NOT NULL DEFAULT '{}',
    evaluated     BOOLEAN     NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_actions_evaluated ON actions(evaluated, created_at);

CREATE TABLE IF NOT EXISTS boundaries (
    id            SERIAL PRIMARY KEY,
    agent_id      TEXT        NOT NULL,
    rule          TEXT        NOT NULL,
    rule_hash     CHAR(64)   NOT NULL,
    active        BOOLEAN     NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(agent_id, rule_hash)
);

CREATE TABLE IF NOT EXISTS violations (
    id            SERIAL PRIMARY KEY,
    agent_id      TEXT        NOT NULL,
    action_id     INTEGER     NOT NULL,
    boundary_id   INTEGER     NOT NULL REFERENCES boundaries(id),
    severity      TEXT        NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    reasoning     TEXT        NOT NULL,
    tx_hash       TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_violations_agent ON violations(agent_id);
CREATE INDEX IF NOT EXISTS idx_violations_action ON violations(action_id);
