CREATE TABLE IF NOT EXISTS tickets (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(200)  NOT NULL,
    description     TEXT,
    status          VARCHAR(20)   NOT NULL DEFAULT 'OPEN',
    priority        VARCHAR(20)   NOT NULL DEFAULT 'MEDIUM',
    assignee        VARCHAR(100),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets(created_at DESC);

-- Seed data (deterministic for fair comparison)
INSERT INTO tickets (title, description, status, priority, assignee, created_at, updated_at)
SELECT
    'Ticket #' || g,
    'Description for ticket number ' || g || '. This is a fixed seed for benchmark reproducibility.',
    (ARRAY['OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED'])[1 + (g % 4)],
    (ARRAY['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'])[1 + (g % 4)],
    'user' || ((g % 10) + 1),
    NOW() - (g || ' minutes')::INTERVAL,
    NOW() - (g || ' minutes')::INTERVAL
FROM generate_series(1, 500) AS g;
