-- ==========================================
-- INFRASTRUCTURE & ORGANIZATIONS
-- ==========================================



-- Organizations and their locations
CREATE TABLE organisations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    street VARCHAR(200) NOT NULL,
    house_number VARCHAR(20) NOT NULL,
    house_number_addition VARCHAR(20) NULL,
    postal_code VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NULL,
    email VARCHAR(100) NULL,
    kvk_number VARCHAR(20) NULL,
    btw_number VARCHAR(20) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app_organization_profile (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    name TEXT NOT NULL DEFAULT '',
    default_timezone TEXT NOT NULL DEFAULT 'Europe/Amsterdam',
    email TEXT NULL,
    phone_number TEXT NULL,
    website TEXT NULL,
    hq_street TEXT NULL,
    hq_house_number TEXT NULL,
    hq_house_number_addition TEXT NULL,
    hq_postal_code TEXT NULL,
    hq_city TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_organization_profile (singleton)
VALUES (TRUE);


-- Create ENUM type for location_type
CREATE TYPE location_type_enum AS ENUM ('care_home', 'office', 'other');
-- Location represents a physical place (care home, apartment building, etc.) for the youth intake
CREATE TABLE location (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    street VARCHAR(200) NOT NULL,
    house_number VARCHAR(20) NOT NULL,
    house_number_addition VARCHAR(20) NULL,
    postal_code VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'Europe/Amsterdam',
    location_type location_type_enum NOT NULL DEFAULT 'other',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- Standard shifts for locations
CREATE TABLE location_shift (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES location(id) ON DELETE CASCADE,
    slot SMALLINT NOT NULL CHECK (slot BETWEEN 1 AND 4),
    shift_name VARCHAR(50) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(location_id, slot),
    UNIQUE(location_id, shift_name)
);

-- Function to insert default shifts for new locations
CREATE OR REPLACE FUNCTION insert_default_shifts()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO location_shift (location_id, slot, shift_name, start_time, end_time)
    VALUES
        (NEW.id, 1, 'Ochtenddienst', TIME '07:30:00', TIME '15:30:00'),
        (NEW.id, 2, 'Avonddienst', TIME '15:00:00', TIME '23:00:00'),
        (NEW.id, 3, 'Slaapdienst of Waakdienst', TIME '23:00:00', TIME '07:30:00');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_insert_default_shifts
AFTER INSERT ON location
FOR EACH ROW
EXECUTE FUNCTION insert_default_shifts();

-- ==========================================
-- USER AUTHENTICATION & PERMISSIONS
-- ==========================================

-- Role templates
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NULL
);

-- System permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    method VARCHAR(255) NOT NULL,
    group_key VARCHAR(100) NOT NULL DEFAULT 'general',
    section_key VARCHAR(100) NOT NULL DEFAULT 'general',
    display_name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TYPE permission_override_effect AS ENUM ('allow', 'deny');

-- Role-to-Permission mapping (template)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- User authentication data
CREATE TABLE custom_user (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    password VARCHAR(128) NOT NULL,
    last_login TIMESTAMPTZ,
    email VARCHAR(254) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    date_joined TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_picture VARCHAR(100),
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret VARCHAR(100),
    two_factor_secret_temp VARCHAR(100),
    recovery_codes TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX custom_user_email_idx ON custom_user(email);
CREATE INDEX custom_user_id_idx ON custom_user(id);

-- Explicit per-user permission exceptions layered on top of role inheritance
CREATE TABLE user_permission_overrides (
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    effect permission_override_effect NOT NULL,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES custom_user(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Track which role templates were given to a user
CREATE TABLE user_roles (
    user_id UUID NOT NULL PRIMARY KEY,
    role_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES custom_user(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

-- Session management for refresh tokens
CREATE TABLE "sessions" (
    "id" uuid PRIMARY KEY,
    "refresh_token" varchar NOT NULL,
    "user_agent" varchar NOT NULL,
    "client_ip" varchar NOT NULL,
    "is_blocked" boolean NOT NULL DEFAULT false,
    "expires_at" timestamptz NOT NULL,
    "created_at" timestamptz NOT NULL,
    "user_id" UUID NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY ("user_id") REFERENCES custom_user("id") ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user ON sessions("user_id");
CREATE INDEX idx_sessions_expires ON sessions("expires_at");
CREATE INDEX idx_sessions_token_blocked ON sessions("refresh_token", "is_blocked");


-- Notification types ENUM
CREATE TYPE notification_type_enum AS ENUM (
    'general'
);

-- Notifications for users
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES custom_user(id) ON DELETE CASCADE,
    type notification_type_enum NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    data JSONB NULL,
    read_at TIMESTAMPTZ NULL DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id_created_at ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_user_id_read_at ON notifications (user_id, read_at);

-- ==========================================
-- FILE MANAGEMENT
-- ==========================================

-- Attachment files
CREATE TABLE attachment_file (
    "uuid" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    "file" VARCHAR(255) NOT NULL,
    "size" INTEGER NOT NULL DEFAULT 0,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    tag VARCHAR(100) NULL DEFAULT '',
    updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX attachment_file_is_used_idx ON attachment_file(is_used);
CREATE INDEX attachment_file_created_idx ON attachment_file(created);

-- Temporary file storage
CREATE TABLE temporary_file (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX temporary_file_uploaded_at_idx ON temporary_file(uploaded_at);

-- ==========================================
-- EMPLOYEE MANAGEMENT
-- ==========================================

-- Shared Gender ENUM
CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other', 'unknown');
-- Employee Contract Type ENUM
CREATE TYPE employee_contract_type_enum AS ENUM ('loondienst', 'ZZP', 'none');

-- Departments (used for employee assignment and handbook templates)
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NULL,
    department_head_employee_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX departments_name_idx ON departments(name);
CREATE INDEX departments_department_head_employee_id_idx ON departments(department_head_employee_id);

-- Employee profile (linked to custom_user)
CREATE TABLE employee_profile (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES custom_user(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    bsn TEXT NOT NULL,
    street TEXT NOT NULL,
    house_number TEXT NOT NULL,
    house_number_addition TEXT NULL,
    postal_code TEXT NOT NULL,
    city TEXT NOT NULL,
    position VARCHAR(100) NULL,
    employee_number VARCHAR(50) NULL,
    employment_number VARCHAR(50) NULL,
    private_email_address VARCHAR(254) NULL,
    work_email_address VARCHAR(254) NULL,
    private_phone_number VARCHAR(100) NULL,
    work_phone_number VARCHAR(100) NULL,
    date_of_birth DATE NULL,
    home_telephone_number VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gender gender_enum NOT NULL,
    location_id UUID NULL REFERENCES location(id) ON DELETE SET NULL,
    department_id UUID NULL REFERENCES departments(id) ON DELETE SET NULL,
    manager_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    has_borrowed BOOLEAN NOT NULL DEFAULT FALSE,
    out_of_service BOOLEAN NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    contract_hours FLOAT NULL DEFAULT 0.0,
    contract_end_date DATE NULL,
    contract_start_date DATE NULL,
    contract_type employee_contract_type_enum NOT NULL DEFAULT 'none',
    contract_rate DECIMAL(10,2) NULL DEFAULT 0.00,
    CONSTRAINT employee_profile_manager_not_self
        CHECK (manager_employee_id IS NULL OR manager_employee_id <> id)
);

CREATE INDEX employee_profile_user_id_idx ON employee_profile(user_id);
CREATE INDEX employee_profile_location_id_idx ON employee_profile(location_id);
CREATE INDEX idx_employee_profile_department_id ON employee_profile(department_id);
CREATE INDEX idx_employee_profile_manager_employee_id ON employee_profile(manager_employee_id);
CREATE INDEX employee_profile_id_desc_idx ON employee_profile(id DESC);
CREATE INDEX idx_employee_profile_is_archived ON employee_profile(is_archived);
CREATE INDEX idx_employee_profile_out_of_service ON employee_profile(out_of_service);

ALTER TABLE departments
    ADD CONSTRAINT departments_department_head_employee_id_fkey
    FOREIGN KEY (department_head_employee_id) REFERENCES employee_profile(id) ON DELETE SET NULL;

-- Employee education records
CREATE TABLE employee_education (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    institution_name VARCHAR(255) NOT NULL,
    degree VARCHAR(100) NOT NULL,
    field_of_study VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX education_employee_id_idx ON employee_education(employee_id);

-- Employee certifications
CREATE TABLE certification (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    issued_by VARCHAR(255) NOT NULL,
    date_issued DATE NOT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX certification_employee_id_idx ON certification(employee_id);

-- Employee work experience
CREATE TABLE employee_experience (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    job_title VARCHAR(255) NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX experience_employee_id_idx ON employee_experience(employee_id);

-- ==========================================
-- EMPLOYEE HANDBOOKS (ONBOARDING)
-- ==========================================

CREATE TYPE handbook_step_kind_enum AS ENUM ('content', 'ack', 'link', 'quiz');
CREATE TYPE handbook_assignment_status_enum AS ENUM ('not_started', 'in_progress', 'completed', 'waived');
CREATE TYPE handbook_step_status_enum AS ENUM ('pending', 'completed', 'skipped');
CREATE TYPE handbook_template_status_enum AS ENUM ('draft', 'published', 'archived');
CREATE TYPE handbook_assignment_event_enum AS ENUM ('assigned', 'reassigned', 'waived', 'started', 'completed');

-- A template is department-specific and can be versioned. Assignments point to a specific version.
CREATE TABLE handbook_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NULL,
    version INT NOT NULL,
    status handbook_template_status_enum NOT NULL DEFAULT 'draft',
    created_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    published_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    published_at TIMESTAMPTZ NULL,
    archived_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (department_id, version)
);

-- Enforce at most one published template per department.
CREATE UNIQUE INDEX handbook_templates_one_published_per_department
    ON handbook_templates(department_id)
    WHERE status = 'published';

-- Enforce at most one draft template per department.
CREATE UNIQUE INDEX handbook_templates_one_draft_per_department
    ON handbook_templates(department_id)
    WHERE status = 'draft';

CREATE INDEX idx_handbook_templates_department_id ON handbook_templates(department_id);

CREATE TABLE handbook_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES handbook_templates(id) ON DELETE CASCADE,
    sort_order INT NOT NULL CHECK (sort_order > 0),
    kind handbook_step_kind_enum NOT NULL DEFAULT 'content',
    title TEXT NOT NULL,
    body TEXT NULL,
    content JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_required BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (template_id, sort_order)
);

CREATE INDEX idx_handbook_steps_template_sort ON handbook_steps(template_id, sort_order);

CREATE TABLE employee_handbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES handbook_templates(id) ON DELETE RESTRICT,
    template_version INT NOT NULL,
    assigned_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    status handbook_assignment_status_enum NOT NULL DEFAULT 'not_started',
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    due_at TIMESTAMPTZ NULL
);

-- At most one active handbook (not_started/in_progress) per employee.
CREATE UNIQUE INDEX employee_handbooks_one_active_per_employee
    ON employee_handbooks(employee_id)
    WHERE status IN ('not_started', 'in_progress');

CREATE INDEX idx_employee_handbooks_employee_assigned_at ON employee_handbooks(employee_id, assigned_at DESC);

CREATE TABLE employee_handbook_assignment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_handbook_id UUID NULL REFERENCES employee_handbooks(id) ON DELETE SET NULL,
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES handbook_templates(id) ON DELETE RESTRICT,
    template_version INT NOT NULL,
    event handbook_assignment_event_enum NOT NULL,
    actor_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_employee_handbook_assignment_history_employee_created_at
    ON employee_handbook_assignment_history(employee_id, created_at DESC);

CREATE TABLE employee_handbook_step_progress (
    employee_handbook_id UUID NOT NULL REFERENCES employee_handbooks(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES handbook_steps(id) ON DELETE RESTRICT,
    status handbook_step_status_enum NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    response JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (employee_handbook_id, step_id)
);

CREATE INDEX idx_employee_handbook_step_progress_handbook_id
    ON employee_handbook_step_progress(employee_handbook_id);



-- ==========================================
-- SCHEDULING & APPOINTMENTS
-- ==========================================

-- Employee schedules
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id),
    location_id UUID NOT NULL REFERENCES location(id),
    location_shift_id UUID NULL REFERENCES location_shift(id),
    shift_name_snapshot VARCHAR(50) NULL,
    shift_start_time_snapshot TIME NULL,
    shift_end_time_snapshot TIME NULL,
    is_custom BOOLEAN NOT NULL DEFAULT FALSE,
    start_datetime TIMESTAMPTZ NOT NULL,
    end_datetime TIMESTAMPTZ NOT NULL,
    created_by_employee_id UUID NOT NULL REFERENCES employee_profile(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_timeframe CHECK (end_datetime > start_datetime),
    CONSTRAINT schedules_custom_shift_link_consistency CHECK (
        (is_custom = TRUE AND location_shift_id IS NULL)
        OR (is_custom = FALSE AND location_shift_id IS NOT NULL)
    ),
    CONSTRAINT schedules_shift_snapshot_consistency CHECK (
        (location_shift_id IS NULL
         AND shift_name_snapshot IS NULL
         AND shift_start_time_snapshot IS NULL
         AND shift_end_time_snapshot IS NULL)
        OR
        (location_shift_id IS NOT NULL
         AND shift_name_snapshot IS NOT NULL
         AND shift_start_time_snapshot IS NOT NULL
         AND shift_end_time_snapshot IS NOT NULL)
    )
);

-- ==========================================
-- LATE ARRIVALS
-- ==========================================

CREATE TABLE late_arrivals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    created_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    arrival_date DATE NOT NULL,
    arrival_time TIME NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT late_arrivals_unique_schedule UNIQUE (schedule_id)
);

CREATE INDEX idx_late_arrivals_employee_id ON late_arrivals(employee_id);
CREATE INDEX idx_late_arrivals_arrival_date_desc ON late_arrivals(arrival_date DESC);
CREATE INDEX idx_late_arrivals_employee_date ON late_arrivals(employee_id, arrival_date DESC);

CREATE TYPE shift_swap_status_enum AS ENUM (
    'pending_recipient',
    'recipient_rejected',
    'pending_admin',
    'admin_rejected',
    'confirmed',
    'cancelled',
    'expired'
);

CREATE TABLE shift_swap_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    recipient_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    requester_schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    recipient_schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    status shift_swap_status_enum NOT NULL DEFAULT 'pending_recipient',
    requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recipient_responded_at TIMESTAMPTZ NULL,
    admin_decided_at TIMESTAMPTZ NULL,
    recipient_response_note TEXT NULL,
    admin_decision_note TEXT NULL,
    admin_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT shift_swap_requester_not_recipient CHECK (requester_employee_id <> recipient_employee_id),
    CONSTRAINT shift_swap_schedule_pair_not_same CHECK (requester_schedule_id <> recipient_schedule_id)
);

CREATE INDEX idx_shift_swap_requests_requester_employee_id ON shift_swap_requests(requester_employee_id);
CREATE INDEX idx_shift_swap_requests_recipient_employee_id ON shift_swap_requests(recipient_employee_id);
CREATE INDEX idx_shift_swap_requests_status ON shift_swap_requests(status);
CREATE INDEX idx_shift_swap_requests_requested_at_desc ON shift_swap_requests(requested_at DESC);
CREATE INDEX idx_shift_swap_requests_expires_at ON shift_swap_requests(expires_at);

CREATE UNIQUE INDEX uq_shift_swap_active_requester_schedule
    ON shift_swap_requests(requester_schedule_id)
    WHERE status IN ('pending_recipient', 'pending_admin');

CREATE UNIQUE INDEX uq_shift_swap_active_recipient_schedule
    ON shift_swap_requests(recipient_schedule_id)
    WHERE status IN ('pending_recipient', 'pending_admin');

CREATE OR REPLACE FUNCTION enforce_shift_swap_active_schedule_uniqueness()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status IN ('pending_recipient', 'pending_admin') THEN
        IF EXISTS (
            SELECT 1
            FROM shift_swap_requests ssr
            WHERE ssr.id <> NEW.id
              AND ssr.status IN ('pending_recipient', 'pending_admin')
              AND (
                ssr.requester_schedule_id IN (NEW.requester_schedule_id, NEW.recipient_schedule_id)
                OR ssr.recipient_schedule_id IN (NEW.requester_schedule_id, NEW.recipient_schedule_id)
              )
        ) THEN
            RAISE EXCEPTION 'one of the schedules is already in an active swap request'
                USING ERRCODE = '23505', CONSTRAINT = 'uq_shift_swap_active_schedule_any';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_shift_swap_active_schedule_uniqueness
BEFORE INSERT OR UPDATE OF requester_schedule_id, recipient_schedule_id, status
ON shift_swap_requests
FOR EACH ROW
EXECUTE FUNCTION enforce_shift_swap_active_schedule_uniqueness();

-- ==========================================
-- LEAVE REQUESTS
-- ==========================================

CREATE TYPE leave_request_type_enum AS ENUM (
    'vacation',
    'personal',
    'sick',
    'pregnancy',
    'unpaid',
    'other'
);

CREATE TYPE leave_request_status_enum AS ENUM (
    'pending',
    'approved',
    'rejected',
    'cancelled',
    'expired'
);

CREATE TABLE leave_policies (
    leave_type leave_request_type_enum PRIMARY KEY,
    requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    deducts_balance BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO leave_policies (leave_type, requires_approval, deducts_balance, is_active) VALUES
    ('vacation', TRUE, TRUE, TRUE),
    ('personal', TRUE, TRUE, TRUE),
    ('sick', FALSE, FALSE, TRUE),
    ('pregnancy', FALSE, FALSE, TRUE),
    ('unpaid', TRUE, FALSE, TRUE),
    ('other', TRUE, FALSE, TRUE);

CREATE TABLE leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    year INT NOT NULL,
    legal_total_days INT NOT NULL DEFAULT 0,
    extra_total_days INT NOT NULL DEFAULT 0,
    legal_used_days INT NOT NULL DEFAULT 0,
    extra_used_days INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_balances_unique_employee_year UNIQUE (employee_id, year),
    CONSTRAINT leave_balances_non_negative CHECK (
        legal_total_days >= 0
        AND extra_total_days >= 0
        AND legal_used_days >= 0
        AND extra_used_days >= 0
    ),
    CONSTRAINT leave_balances_usage_not_exceed_total CHECK (
        legal_used_days <= legal_total_days
        AND extra_used_days <= extra_total_days
    )
);

CREATE INDEX idx_leave_balances_employee_year ON leave_balances(employee_id, year);

CREATE TABLE leave_balance_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    leave_balance_id UUID NOT NULL REFERENCES leave_balances(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    year INT NOT NULL,
    legal_days_delta INT NOT NULL DEFAULT 0,
    extra_days_delta INT NOT NULL DEFAULT 0,
    reason TEXT NOT NULL,
    adjusted_by_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE RESTRICT,
    legal_total_days_before INT NOT NULL,
    extra_total_days_before INT NOT NULL,
    legal_total_days_after INT NOT NULL,
    extra_total_days_after INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_balance_adjustments_non_zero_delta CHECK (
        legal_days_delta <> 0 OR extra_days_delta <> 0
    )
);

CREATE INDEX idx_leave_balance_adjustments_employee_year_created_at
ON leave_balance_adjustments(employee_id, year, created_at DESC);

CREATE OR REPLACE FUNCTION initialize_leave_balance_on_employee_insert()
RETURNS TRIGGER AS $$
DECLARE
    computed_legal_days INT;
    current_year INT;
BEGIN
    current_year := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    computed_legal_days := GREATEST(0, ROUND(COALESCE(NEW.contract_hours, 0)::numeric / 2.0)::INT);

    INSERT INTO leave_balances (
        employee_id,
        year,
        legal_total_days,
        extra_total_days,
        legal_used_days,
        extra_used_days
    ) VALUES (
        NEW.id,
        current_year,
        computed_legal_days,
        0,
        0,
        0
    )
    ON CONFLICT (employee_id, year) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_initialize_leave_balance_on_employee_insert
AFTER INSERT ON employee_profile
FOR EACH ROW
EXECUTE FUNCTION initialize_leave_balance_on_employee_insert();

CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    created_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    leave_type leave_request_type_enum NOT NULL,
    status leave_request_status_enum NOT NULL DEFAULT 'pending',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NULL,
    decision_note TEXT NULL,
    decided_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    decided_at TIMESTAMPTZ NULL,
    cancelled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_requests_date_order CHECK (end_date >= start_date)
);

CREATE INDEX idx_leave_requests_employee_id ON leave_requests(employee_id);
CREATE INDEX idx_leave_requests_status ON leave_requests(status);
CREATE INDEX idx_leave_requests_leave_type ON leave_requests(leave_type);
CREATE INDEX idx_leave_requests_requested_at_desc ON leave_requests(requested_at DESC);
CREATE INDEX idx_leave_requests_employee_status ON leave_requests(employee_id, status);

CREATE TYPE calendar_event_kind_enum AS ENUM ('appointment', 'reminder');
CREATE TYPE calendar_event_status_enum AS ENUM ('confirmed', 'cancelled');
-- Work approval status for appointments (hours are counted/billed only after admin approval)
CREATE TYPE calendar_event_work_approval_status_enum AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE attendee_response_enum AS ENUM ('needs_action', 'accepted', 'declined', 'tentative');
CREATE TYPE reminder_channel_enum AS ENUM ('in_app');

CREATE TABLE calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    created_by_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    kind calendar_event_kind_enum NOT NULL,
    status calendar_event_status_enum NOT NULL DEFAULT 'confirmed',
    work_approval_status calendar_event_work_approval_status_enum NOT NULL DEFAULT 'pending',
    work_approved_by UUID NULL REFERENCES custom_user(id) ON DELETE SET NULL,
    work_approved_at TIMESTAMPTZ NULL,
    work_rejected_by UUID NULL REFERENCES custom_user(id) ON DELETE SET NULL,
    work_rejected_at TIMESTAMPTZ NULL,
    work_rejection_reason TEXT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NULL,
    location TEXT NULL,
    color VARCHAR(20) NULL,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    rrule TEXT NULL,
    recurring_event_id UUID NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    recurrence_id TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_events_time_order CHECK (end_at > start_at),
    CONSTRAINT calendar_events_exception_shape CHECK (
        (recurring_event_id IS NULL AND recurrence_id IS NULL)
        OR
        (recurring_event_id IS NOT NULL AND recurrence_id IS NOT NULL)
    ),
    CONSTRAINT calendar_events_exception_no_rrule CHECK (
        recurring_event_id IS NULL OR rrule IS NULL
    )
);

CREATE INDEX idx_calendar_events_organizer_start ON calendar_events(organizer_employee_id, start_at);
CREATE INDEX idx_calendar_events_start ON calendar_events(start_at);
CREATE UNIQUE INDEX uq_calendar_events_exception ON calendar_events(recurring_event_id, recurrence_id)
    WHERE recurring_event_id IS NOT NULL;

-- If appointment times change, previously approved hours should be re-approved.
CREATE OR REPLACE FUNCTION calendar_event_reset_work_approval_on_time_change()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.kind = 'appointment' THEN
        IF (NEW.start_at IS DISTINCT FROM OLD.start_at) OR (NEW.end_at IS DISTINCT FROM OLD.end_at) THEN
            NEW.work_approval_status := 'pending';
            NEW.work_approved_by := NULL;
            NEW.work_approved_at := NULL;
            NEW.work_rejected_by := NULL;
            NEW.work_rejected_at := NULL;
            NEW.work_rejection_reason := NULL;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_calendar_event_reset_work_approval
BEFORE UPDATE ON calendar_events
FOR EACH ROW
EXECUTE FUNCTION calendar_event_reset_work_approval_on_time_change();

CREATE TABLE calendar_event_attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    email TEXT NULL,
    response attendee_response_enum NOT NULL DEFAULT 'needs_action',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_event_attendees_one_target CHECK (
        ((employee_id IS NOT NULL)::INT + (email IS NOT NULL)::INT) = 1
    )
);

CREATE UNIQUE INDEX uq_event_attendee_employee ON calendar_event_attendees(event_id, employee_id)
    WHERE employee_id IS NOT NULL;
CREATE UNIQUE INDEX uq_event_attendee_email ON calendar_event_attendees(event_id, email)
    WHERE email IS NOT NULL;
CREATE INDEX idx_event_attendees_employee ON calendar_event_attendees(employee_id);

CREATE TABLE calendar_event_reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    channel reminder_channel_enum NOT NULL DEFAULT 'in_app',
    minutes_before INT NULL,
    remind_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT calendar_event_reminders_one_mode CHECK (
        (minutes_before IS NOT NULL) <> (remind_at IS NOT NULL)
    ),
    CONSTRAINT calendar_event_reminders_minutes_positive CHECK (
        minutes_before IS NULL OR minutes_before >= 0
    )
);

CREATE INDEX idx_event_reminders_event ON calendar_event_reminders(event_id);



-- Helper function to get current employee ID
CREATE OR REPLACE FUNCTION get_current_employee_id() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('myapp.current_employee_id', true)::UUID;
EXCEPTION
    WHEN OTHERS THEN RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Check if current employee is Admin
CREATE OR REPLACE FUNCTION is_admin() RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        JOIN employee_profile ep ON ep.user_id = ur.user_id
        WHERE ep.id = get_current_employee_id()
        AND r.name = 'admin'
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Check if current employee is Coordinator
CREATE OR REPLACE FUNCTION is_coordinator() RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        JOIN employee_profile ep ON ep.user_id = ur.user_id
        WHERE ep.id = get_current_employee_id()
        AND r.name = 'coordinator'
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

