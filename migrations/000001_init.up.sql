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
    name VARCHAR(255) NOT NULL UNIQUE,
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

-- Bootstrap default admin role
INSERT INTO roles (name, description)
VALUES ('admin', 'System administrator with full access')
ON CONFLICT (name) DO NOTHING;

-- Bootstrap known API permissions (code-derived)
WITH seeded(name, sort_order) AS (
    VALUES
        ('EMPLOYEE.CREATE', 10),
        ('EMPLOYEE.DELETE', 20),
        ('EMPLOYEE.UPDATE', 30),
        ('EMPLOYEE.VIEW', 40),
        ('HANDBOOK.ASSIGN', 50),
        ('HANDBOOK.SELF.UPDATE', 60),
        ('HANDBOOK.SELF.VIEW', 70),
        ('HANDBOOK.STEP.CREATE', 80),
        ('HANDBOOK.STEP.DELETE', 90),
        ('HANDBOOK.STEP.UPDATE', 100),
        ('HANDBOOK.STEP.VIEW', 110),
        ('HANDBOOK.TEMPLATE.CREATE', 120),
        ('HANDBOOK.TEMPLATE.PUBLISH', 130),
        ('HANDBOOK.TEMPLATE.UPDATE', 140),
        ('HANDBOOK.TEMPLATE.VIEW', 150),
        ('LEAVE.BALANCE.ADJUST', 160),
        ('LEAVE.BALANCE.VIEW', 170),
        ('LEAVE.BALANCE.VIEW_ALL', 180),
        ('LEAVE.REQUEST.CREATE', 190),
        ('LEAVE.REQUEST.DECIDE', 200),
        ('LEAVE.REQUEST.UPDATE', 210),
        ('LEAVE.REQUEST.UPDATE_ALL', 220),
        ('LEAVE.REQUEST.VIEW', 230),
        ('LEAVE.REQUEST.VIEW_ALL', 240),
        ('LOCATION.CREATE', 250),
        ('LOCATION.DELETE', 260),
        ('LOCATION.UPDATE', 270),
        ('LOCATION.VIEW', 280),
        ('PAYOUT.REQUEST.CREATE', 290),
        ('PAYOUT.REQUEST.DECIDE', 300),
        ('PAYOUT.REQUEST.MARK_PAID', 310),
        ('PAYOUT.REQUEST.VIEW', 320),
        ('PAYOUT.REQUEST.VIEW_ALL', 330),
        ('PAY_PERIOD.CLOSE', 340),
        ('PAY_PERIOD.MARK_PAID', 350),
        ('PAY_PERIOD.MONTH_SUMMARY_VIEW', 360),
        ('PAY_PERIOD.VIEW_ALL', 370),
        ('ROLE.VIEW', 375),
        ('PERFORMANCE.ASSESSMENT.CREATE', 560),
        ('PERFORMANCE.ASSESSMENT.VIEW', 570),
        ('PERFORMANCE.ASSESSMENT.VIEW_ALL', 580),
        ('PERFORMANCE.ASSESSMENT.DELETE', 590),
        ('PERFORMANCE.WORK_ASSIGNMENT.VIEW', 600),
        ('PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL', 610),
        ('PERFORMANCE.WORK_ASSIGNMENT.DECIDE', 620),
        ('PERFORMANCE.UPCOMING.INVITE', 630),
        ('PERFORMANCE.STATS.VIEW', 640),
        ('SCHEDULE.CREATE', 380),
        ('SCHEDULE.DELETE', 390),
        ('SCHEDULE.UPDATE', 400),
        ('SCHEDULE.VIEW', 410),
        ('SCHEDULE_SWAP.APPROVE', 420),
        ('SCHEDULE_SWAP.REQUEST', 430),
        ('SCHEDULE_SWAP.RESPOND', 440),
        ('SCHEDULE_SWAP.VIEW', 450),
        ('SHIFT.CREATE', 460),
        ('SHIFT.DELETE', 470),
        ('SHIFT.UPDATE', 480),
        ('SHIFT.VIEW', 490),
        ('TIME_ENTRY.CREATE', 500),
        ('TIME_ENTRY.CREATE_ALL', 510),
        ('TIME_ENTRY.UPDATE_ALL', 520),
        ('TIME_ENTRY.VIEW', 530),
        ('TIME_ENTRY.VIEW_ALL', 540),
        ('TIME_ENTRY.DECIDE', 550)
)
INSERT INTO permissions (
    name,
    resource,
    method,
    group_key,
    section_key,
    display_name,
    description,
    sort_order
)
SELECT
    s.name,
    split_part(s.name, '.', 1) AS resource,
    CASE
        WHEN strpos(s.name, '.') > 0 THEN substr(s.name, strpos(s.name, '.') + 1)
        ELSE s.name
    END AS method,
    lower(split_part(s.name, '.', 1)) AS group_key,
    lower(COALESCE(NULLIF(split_part(s.name, '.', 2), ''), 'general')) AS section_key,
    initcap(replace(lower(s.name), '.', ' ')) AS display_name,
    NULL,
    s.sort_order
FROM seeded s
ON CONFLICT (name) DO UPDATE SET
    resource = EXCLUDED.resource,
    method = EXCLUDED.method,
    group_key = EXCLUDED.group_key,
    section_key = EXCLUDED.section_key,
    display_name = CASE
        WHEN permissions.display_name = '' THEN EXCLUDED.display_name
        ELSE permissions.display_name
    END,
    sort_order = EXCLUDED.sort_order;

-- Grant all seeded permissions to admin
WITH admin_role AS (
    SELECT id
    FROM roles
    WHERE name = 'admin'
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT ar.id, p.id
FROM admin_role ar
CROSS JOIN permissions p
WHERE p.name IN (
    'EMPLOYEE.CREATE',
    'EMPLOYEE.DELETE',
    'EMPLOYEE.UPDATE',
    'EMPLOYEE.VIEW',
    'HANDBOOK.ASSIGN',
    'HANDBOOK.SELF.UPDATE',
    'HANDBOOK.SELF.VIEW',
    'HANDBOOK.STEP.CREATE',
    'HANDBOOK.STEP.DELETE',
    'HANDBOOK.STEP.UPDATE',
    'HANDBOOK.STEP.VIEW',
    'HANDBOOK.TEMPLATE.CREATE',
    'HANDBOOK.TEMPLATE.PUBLISH',
    'HANDBOOK.TEMPLATE.UPDATE',
    'HANDBOOK.TEMPLATE.VIEW',
    'LEAVE.BALANCE.ADJUST',
    'LEAVE.BALANCE.VIEW',
    'LEAVE.BALANCE.VIEW_ALL',
    'LEAVE.REQUEST.CREATE',
    'LEAVE.REQUEST.DECIDE',
    'LEAVE.REQUEST.UPDATE',
    'LEAVE.REQUEST.UPDATE_ALL',
    'LEAVE.REQUEST.VIEW',
    'LEAVE.REQUEST.VIEW_ALL',
    'LOCATION.CREATE',
    'LOCATION.DELETE',
    'LOCATION.UPDATE',
    'LOCATION.VIEW',
    'PAYOUT.REQUEST.CREATE',
    'PAYOUT.REQUEST.DECIDE',
    'PAYOUT.REQUEST.MARK_PAID',
    'PAYOUT.REQUEST.VIEW',
    'PAYOUT.REQUEST.VIEW_ALL',
    'PAY_PERIOD.CLOSE',
    'PAY_PERIOD.MARK_PAID',
    'PAY_PERIOD.MONTH_SUMMARY_VIEW',
    'PAY_PERIOD.VIEW_ALL',
    'ROLE.VIEW',
    'PERFORMANCE.ASSESSMENT.CREATE',
    'PERFORMANCE.ASSESSMENT.VIEW',
    'PERFORMANCE.ASSESSMENT.VIEW_ALL',
    'PERFORMANCE.ASSESSMENT.DELETE',
    'PERFORMANCE.WORK_ASSIGNMENT.VIEW',
    'PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL',
    'PERFORMANCE.WORK_ASSIGNMENT.DECIDE',
    'PERFORMANCE.UPCOMING.INVITE',
    'PERFORMANCE.STATS.VIEW',
    'SCHEDULE.CREATE',
    'SCHEDULE.DELETE',
    'SCHEDULE.UPDATE',
    'SCHEDULE.VIEW',
    'SCHEDULE_SWAP.APPROVE',
    'SCHEDULE_SWAP.REQUEST',
    'SCHEDULE_SWAP.RESPOND',
    'SCHEDULE_SWAP.VIEW',
    'SHIFT.CREATE',
    'SHIFT.DELETE',
    'SHIFT.UPDATE',
    'SHIFT.VIEW',
    'TIME_ENTRY.CREATE',
    'TIME_ENTRY.CREATE_ALL',
    'TIME_ENTRY.UPDATE_ALL',
    'TIME_ENTRY.VIEW',
    'TIME_ENTRY.VIEW_ALL',
    'TIME_ENTRY.DECIDE'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

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
CREATE TYPE irregular_hours_profile_enum AS ENUM ('none', 'roster', 'non_roster');

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
    irregular_hours_profile irregular_hours_profile_enum NOT NULL DEFAULT 'none',
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

CREATE TABLE employee_contract_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    effective_from DATE NOT NULL,
    contract_hours FLOAT NOT NULL DEFAULT 0.0,
    contract_type employee_contract_type_enum NOT NULL DEFAULT 'none',
    contract_rate DECIMAL(10,2) NULL DEFAULT 0.00,
    irregular_hours_profile irregular_hours_profile_enum NOT NULL DEFAULT 'none',
    contract_end_date DATE NULL,
    created_by_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT employee_contract_changes_unique_employee_effective_from UNIQUE (employee_id, effective_from),
    CONSTRAINT employee_contract_changes_contract_hours_non_negative CHECK (contract_hours >= 0)
);

CREATE INDEX idx_employee_contract_changes_employee_effective_from_desc
ON employee_contract_changes(employee_id, effective_from DESC);

CREATE TABLE national_holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country_code TEXT NOT NULL,
    holiday_date DATE NOT NULL,
    name TEXT NOT NULL,
    is_national BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT national_holidays_country_code_not_blank CHECK (btrim(country_code) <> ''),
    CONSTRAINT national_holidays_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT national_holidays_unique_country_date UNIQUE (country_code, holiday_date)
);

CREATE INDEX idx_national_holidays_country_date
ON national_holidays(country_code, holiday_date);

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
-- TIME ENTRIES
-- ==========================================

CREATE TYPE time_entry_status_enum AS ENUM ('draft', 'submitted', 'approved', 'rejected');
CREATE TYPE time_entry_hour_type_enum AS ENUM ('normal', 'overtime', 'travel', 'leave', 'sick', 'training');

CREATE TABLE time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    schedule_id UUID NULL REFERENCES schedules(id) ON DELETE SET NULL,
    entry_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    break_minutes INTEGER NOT NULL DEFAULT 0 CHECK (break_minutes >= 0),
    hour_type time_entry_hour_type_enum NOT NULL DEFAULT 'normal',
    project_name TEXT,
    project_number TEXT,
    client_name TEXT,
    activity_category TEXT,
    activity_description TEXT,
    status time_entry_status_enum NOT NULL DEFAULT 'draft',
    submitted_at TIMESTAMPTZ NULL,
    approved_at TIMESTAMPTZ NULL,
    approved_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    rejection_reason TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT time_entries_non_zero_duration CHECK (start_time <> end_time)
);

CREATE INDEX idx_time_entries_employee_date ON time_entries(employee_id, entry_date DESC);
CREATE INDEX idx_time_entries_status ON time_entries(status);
CREATE INDEX idx_time_entries_schedule_id ON time_entries(schedule_id);

CREATE TABLE time_entry_update_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    time_entry_id UUID NOT NULL REFERENCES time_entries(id) ON DELETE CASCADE,
    admin_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE RESTRICT,
    admin_update_note TEXT NOT NULL CHECK (btrim(admin_update_note) <> ''),
    before_snapshot JSONB NOT NULL,
    after_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_time_entry_update_audits_entry_created_at
ON time_entry_update_audits(time_entry_id, created_at DESC);

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

CREATE TYPE payout_request_status_enum AS ENUM (
    'pending',
    'approved',
    'rejected',
    'paid'
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
    legal_total_hours INT NOT NULL DEFAULT 0,
    extra_total_hours INT NOT NULL DEFAULT 0,
    legal_used_hours INT NOT NULL DEFAULT 0,
    extra_used_hours INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_balances_unique_employee_year UNIQUE (employee_id, year),
    CONSTRAINT leave_balances_non_negative CHECK (
        legal_total_hours >= 0
        AND extra_total_hours >= 0
        AND legal_used_hours >= 0
        AND extra_used_hours >= 0
    ),
    CONSTRAINT leave_balances_usage_not_exceed_total CHECK (
        legal_used_hours <= legal_total_hours
        AND extra_used_hours <= extra_total_hours
    )
);

CREATE INDEX idx_leave_balances_employee_year ON leave_balances(employee_id, year);

CREATE TABLE leave_balance_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    leave_balance_id UUID NOT NULL REFERENCES leave_balances(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    year INT NOT NULL,
    legal_hours_delta INT NOT NULL DEFAULT 0,
    extra_hours_delta INT NOT NULL DEFAULT 0,
    reason TEXT NOT NULL,
    adjusted_by_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE RESTRICT,
    legal_total_hours_before INT NOT NULL,
    extra_total_hours_before INT NOT NULL,
    legal_total_hours_after INT NOT NULL,
    extra_total_hours_after INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_balance_adjustments_non_zero_delta CHECK (
        legal_hours_delta <> 0 OR extra_hours_delta <> 0
    )
);

CREATE INDEX idx_leave_balance_adjustments_employee_year_created_at
ON leave_balance_adjustments(employee_id, year, created_at DESC);

CREATE OR REPLACE FUNCTION calculate_legal_leave_hours(p_employee_id UUID, p_year INT)
RETURNS INT AS $$
DECLARE
    computed_legal_hours INT := 0;
    has_history BOOLEAN := FALSE;
    profile_contract_type employee_contract_type_enum;
    profile_contract_hours FLOAT;
BEGIN
    SELECT EXISTS(
        SELECT 1
        FROM employee_contract_changes ecc
        WHERE ecc.employee_id = p_employee_id
    ) INTO has_history;

    IF has_history THEN
        SELECT
            GREATEST(
                0,
                ROUND(COALESCE(SUM(
                    CASE
                        WHEN segments.contract_type <> 'loondienst' OR segments.contract_hours <= 0 THEN 0
                        ELSE (
                            (segments.contract_hours * 4.0) * (
                                (
                                    LEAST(segments.segment_end, make_date(p_year, 12, 31)) -
                                    GREATEST(segments.effective_from, make_date(p_year, 1, 1)) + 1
                                )::numeric /
                                (make_date(p_year + 1, 1, 1) - make_date(p_year, 1, 1))::numeric
                            )
                        )
                    END
                ), 0)::numeric)::INT
            )
        INTO computed_legal_hours
        FROM (
            SELECT
                ecc.effective_from,
                COALESCE(
                    LEAST(
                        COALESCE(ecc.contract_end_date, make_date(p_year, 12, 31)),
                        (
                            LEAD(ecc.effective_from) OVER (
                                PARTITION BY ecc.employee_id
                                ORDER BY ecc.effective_from
                            ) - INTERVAL '1 day'
                        )::DATE
                    ),
                    COALESCE(ecc.contract_end_date, make_date(p_year, 12, 31))
                ) AS segment_end,
                ecc.contract_hours,
                ecc.contract_type
            FROM employee_contract_changes ecc
            WHERE ecc.employee_id = p_employee_id
        ) AS segments
        WHERE segments.segment_end >= make_date(p_year, 1, 1)
          AND segments.effective_from <= make_date(p_year, 12, 31);

        RETURN COALESCE(computed_legal_hours, 0);
    END IF;

    SELECT
        ep.contract_type,
        COALESCE(ep.contract_hours, 0)
    INTO
        profile_contract_type,
        profile_contract_hours
    FROM employee_profile ep
    WHERE ep.id = p_employee_id;

    IF profile_contract_type IS NULL OR profile_contract_type <> 'loondienst' THEN
        RETURN 0;
    END IF;

    IF profile_contract_hours > 0 THEN
        RETURN GREATEST(0, ROUND((profile_contract_hours * 4)::numeric)::INT);
    END IF;

    SELECT
        GREATEST(
            0,
            ROUND(
                COALESCE(
                    SUM(
                        GREATEST(
                            0,
                            (
                                CASE
                                    WHEN te.end_time > te.start_time THEN
                                        EXTRACT(EPOCH FROM te.end_time) - EXTRACT(EPOCH FROM te.start_time)
                                    ELSE
                                        EXTRACT(EPOCH FROM te.end_time) + 86400 - EXTRACT(EPOCH FROM te.start_time)
                                END
                            ) / 3600.0 - (te.break_minutes::numeric / 60.0)
                        )
                    ) * (4.0 / 52.0),
                    0
                )::numeric
            )::INT
        )
    INTO computed_legal_hours
    FROM time_entries te
    WHERE te.employee_id = p_employee_id
      AND te.status = 'approved'::time_entry_status_enum
      AND te.hour_type IN (
          'normal'::time_entry_hour_type_enum,
          'overtime'::time_entry_hour_type_enum,
          'travel'::time_entry_hour_type_enum,
          'training'::time_entry_hour_type_enum
      )
      AND te.entry_date >= make_date(p_year, 1, 1)
      AND te.entry_date < make_date(p_year + 1, 1, 1);

    RETURN COALESCE(computed_legal_hours, 0);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION initialize_leave_balance_on_employee_insert()
RETURNS TRIGGER AS $$
DECLARE
    computed_legal_hours INT;
    current_year INT;
BEGIN
    current_year := EXTRACT(YEAR FROM CURRENT_DATE)::INT;

    IF NEW.contract_type <> 'loondienst' OR COALESCE(NEW.contract_hours, 0) <= 0 THEN
        RETURN NEW;
    END IF;

    computed_legal_hours := calculate_legal_leave_hours(NEW.id, current_year);

    INSERT INTO leave_balances (
        employee_id,
        year,
        legal_total_hours,
        extra_total_hours,
        legal_used_hours,
        extra_used_hours
    ) VALUES (
        NEW.id,
        current_year,
        computed_legal_hours,
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

CREATE TABLE leave_payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    created_by_employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE RESTRICT,
    requested_hours INT NOT NULL,
    balance_year INT NOT NULL,
    hourly_rate DECIMAL(10,2) NOT NULL,
    gross_amount DECIMAL(12,2) NOT NULL,
    salary_month DATE NULL,
    status payout_request_status_enum NOT NULL DEFAULT 'pending',
    request_note TEXT NULL,
    decision_note TEXT NULL,
    decided_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    paid_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    decided_at TIMESTAMPTZ NULL,
    paid_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT leave_payout_requests_requested_hours_positive CHECK (requested_hours > 0),
    CONSTRAINT leave_payout_requests_balance_year_range CHECK (balance_year >= 2000 AND balance_year <= 2100),
    CONSTRAINT leave_payout_requests_hourly_rate_positive CHECK (hourly_rate > 0),
    CONSTRAINT leave_payout_requests_gross_amount_non_negative CHECK (gross_amount >= 0),
    CONSTRAINT leave_payout_requests_salary_month_first_day CHECK (
        salary_month IS NULL OR EXTRACT(DAY FROM salary_month) = 1
    )
);

CREATE INDEX idx_leave_payout_requests_employee_requested_at_desc
ON leave_payout_requests(employee_id, requested_at DESC);
CREATE INDEX idx_leave_payout_requests_status_requested_at_desc
ON leave_payout_requests(status, requested_at DESC);
CREATE INDEX idx_leave_payout_requests_balance_year
ON leave_payout_requests(balance_year);

CREATE TYPE pay_period_status_enum AS ENUM ('draft', 'paid');

CREATE TABLE pay_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status pay_period_status_enum NOT NULL DEFAULT 'draft',
    base_gross_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    irregular_gross_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    gross_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ NULL,
    created_by_employee_id UUID NULL REFERENCES employee_profile(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pay_periods_period_order CHECK (period_end >= period_start),
    CONSTRAINT pay_periods_base_gross_non_negative CHECK (base_gross_amount >= 0),
    CONSTRAINT pay_periods_irregular_gross_non_negative CHECK (irregular_gross_amount >= 0),
    CONSTRAINT pay_periods_gross_non_negative CHECK (gross_amount >= 0),
    CONSTRAINT pay_periods_unique_employee_period UNIQUE (employee_id, period_start, period_end)
);

CREATE INDEX idx_pay_periods_employee_period
ON pay_periods(employee_id, period_start DESC, period_end DESC);

CREATE INDEX idx_pay_periods_status_period
ON pay_periods(status, period_start DESC, period_end DESC);

CREATE TABLE pay_period_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pay_period_id UUID NOT NULL REFERENCES pay_periods(id) ON DELETE CASCADE,
    time_entry_id UUID NULL REFERENCES time_entries(id) ON DELETE SET NULL,
    work_date DATE NOT NULL,
    line_type TEXT NOT NULL,
    irregular_hours_profile irregular_hours_profile_enum NOT NULL DEFAULT 'none',
    applied_rate_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    minutes_worked DECIMAL(10,2) NOT NULL DEFAULT 0,
    base_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    premium_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pay_period_line_items_line_type_not_blank CHECK (btrim(line_type) <> ''),
    CONSTRAINT pay_period_line_items_applied_rate_non_negative CHECK (applied_rate_percent >= 0),
    CONSTRAINT pay_period_line_items_minutes_non_negative CHECK (minutes_worked >= 0),
    CONSTRAINT pay_period_line_items_base_non_negative CHECK (base_amount >= 0),
    CONSTRAINT pay_period_line_items_premium_non_negative CHECK (premium_amount >= 0)
);

CREATE INDEX idx_pay_period_line_items_pay_period
ON pay_period_line_items(pay_period_id, work_date ASC, created_at ASC);

ALTER TABLE time_entries
    ADD COLUMN paid_period_id UUID NULL REFERENCES pay_periods(id) ON DELETE SET NULL;

CREATE INDEX idx_time_entries_paid_period_id
ON time_entries(paid_period_id);

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

CREATE TYPE performance_assessment_status_enum AS ENUM ('draft', 'completed');
CREATE TYPE performance_work_assignment_status_enum AS ENUM (
    'open',
    'submitted',
    'approved',
    'revision_needed'
);

CREATE TABLE performance_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    assessment_date DATE NOT NULL,
    total_score NUMERIC(4,2) NULL,
    status performance_assessment_status_enum NOT NULL DEFAULT 'draft',
    notes TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_performance_assessments_employee_date
ON performance_assessments(employee_id, assessment_date DESC);

CREATE INDEX idx_performance_assessments_status
ON performance_assessments(status);

CREATE TABLE performance_assessment_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id UUID NOT NULL REFERENCES performance_assessments(id) ON DELETE CASCADE,
    domain_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    rating NUMERIC(4,2) NOT NULL,
    remarks TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT performance_assessment_scores_rating_range CHECK (rating >= 1 AND rating <= 10),
    CONSTRAINT performance_assessment_scores_unique_item UNIQUE (assessment_id, domain_id, item_id)
);

CREATE INDEX idx_performance_assessment_scores_assessment
ON performance_assessment_scores(assessment_id);

CREATE TABLE performance_work_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id UUID NOT NULL REFERENCES performance_assessments(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employee_profile(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL,
    domain_id TEXT NOT NULL,
    question_text TEXT NOT NULL,
    score NUMERIC(4,2) NOT NULL,
    assignment_description TEXT NOT NULL,
    improvement_notes TEXT NULL,
    expectations TEXT NULL,
    advice TEXT NULL,
    due_date DATE NULL,
    status performance_work_assignment_status_enum NOT NULL DEFAULT 'open',
    submitted_at TIMESTAMPTZ NULL,
    submission_text TEXT NULL,
    feedback TEXT NULL,
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_performance_work_assignments_employee_due
ON performance_work_assignments(employee_id, due_date);

CREATE INDEX idx_performance_work_assignments_status
ON performance_work_assignments(status);



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
