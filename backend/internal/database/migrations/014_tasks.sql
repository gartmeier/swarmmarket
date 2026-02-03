-- Migration 014: Task Protocol
-- Creates tables for capability-linked task execution with schema validation

-- Tasks table: capability-linked unit of work
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Parties
    requester_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    executor_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    capability_id UUID NOT NULL REFERENCES capabilities(id) ON DELETE CASCADE,

    -- Task input (validated against capability input_schema)
    input JSONB NOT NULL,

    -- Task output (validated against capability output_schema on completion)
    output JSONB,

    -- Status machine
    -- Core statuses: pending, accepted, in_progress, delivered, completed, cancelled, failed
    status VARCHAR(30) NOT NULL DEFAULT 'pending',

    -- Current custom status event (from capability.status_events)
    current_event VARCHAR(100),
    current_event_data JSONB,

    -- Callback for status updates
    callback_url VARCHAR(2048),
    callback_secret VARCHAR(64),

    -- Pricing (copied from capability at creation time)
    price_amount DECIMAL(20, 8) NOT NULL,
    price_currency VARCHAR(10) NOT NULL DEFAULT 'USD',

    -- Linked transaction (created when task is accepted)
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,

    -- Error handling
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,

    -- Deadlines and timestamps
    deadline_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for tasks
CREATE INDEX IF NOT EXISTS idx_tasks_requester ON tasks(requester_id);
CREATE INDEX IF NOT EXISTS idx_tasks_executor ON tasks(executor_id);
CREATE INDEX IF NOT EXISTS idx_tasks_capability ON tasks(capability_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_transaction ON tasks(transaction_id) WHERE transaction_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline_at) WHERE deadline_at IS NOT NULL;

-- Task status history (audit log for state transitions)
CREATE TABLE IF NOT EXISTS task_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    from_status VARCHAR(30),
    to_status VARCHAR(30) NOT NULL,
    event VARCHAR(100),
    event_data JSONB,
    changed_by UUID REFERENCES agents(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for task status history
CREATE INDEX IF NOT EXISTS idx_task_history_task ON task_status_history(task_id);
CREATE INDEX IF NOT EXISTS idx_task_history_created ON task_status_history(created_at DESC);

-- Add task_id column to transactions table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'task_id'
    ) THEN
        ALTER TABLE transactions ADD COLUMN task_id UUID REFERENCES tasks(id) ON DELETE SET NULL;
        CREATE INDEX idx_transactions_task ON transactions(task_id) WHERE task_id IS NOT NULL;
    END IF;
END $$;

-- Trigger for updated_at on tasks
CREATE OR REPLACE FUNCTION update_tasks_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tasks_updated ON tasks;
CREATE TRIGGER tasks_updated
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_tasks_timestamp();
