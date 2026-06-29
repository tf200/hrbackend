CREATE TYPE schedule_history_action_enum AS ENUM (
    'created',
    'updated',
    'deleted',
    'employee_assignment_changed'
);

CREATE TABLE schedule_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL,
    location_id UUID NOT NULL REFERENCES location(id) ON DELETE CASCADE,
    action schedule_history_action_enum NOT NULL,
    actor_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    old_values JSONB NULL,
    new_values JSONB NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_schedule_history_schedule_created_at
ON schedule_history(schedule_id, created_at DESC);

CREATE INDEX idx_schedule_history_location_created_at
ON schedule_history(location_id, created_at DESC);

CREATE INDEX idx_schedule_history_actor_created_at
ON schedule_history(actor_employee_id, created_at DESC);
