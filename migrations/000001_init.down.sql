-- ==========================================
-- DOWN MIGRATION - Rollback all changes
-- ==========================================

-- ==========================================
-- DROP FUNCTIONS & TRIGGERS (reverse dependency order)
-- ==========================================

-- Drop trigger and function for shift swap active schedule uniqueness
DROP TRIGGER IF EXISTS trg_shift_swap_active_schedule_uniqueness ON shift_swap_requests;
DROP FUNCTION IF EXISTS enforce_shift_swap_active_schedule_uniqueness();

-- Drop trigger and function for calendar event work approval reset
DROP TRIGGER IF EXISTS trigger_calendar_event_reset_work_approval ON calendar_events;
DROP FUNCTION IF EXISTS calendar_event_reset_work_approval_on_time_change();

-- Drop trigger and function for leave balance initialization
DROP TRIGGER IF EXISTS trigger_initialize_leave_balance_on_employee_insert ON employee_profile;
DROP FUNCTION IF EXISTS initialize_leave_balance_on_employee_insert();
DROP FUNCTION IF EXISTS calculate_legal_leave_hours(UUID, INT);

-- Drop trigger and function for default shifts on location
DROP TRIGGER IF EXISTS trigger_insert_default_shifts ON location;
DROP FUNCTION IF EXISTS insert_default_shifts();

-- Drop helper functions
DROP FUNCTION IF EXISTS get_current_employee_id();
DROP FUNCTION IF EXISTS is_admin();
DROP FUNCTION IF EXISTS is_coordinator();

-- ==========================================
-- DROP TABLES (reverse dependency order)
-- ==========================================

-- Calendar tables
DROP TABLE IF EXISTS calendar_event_reminders;
DROP TABLE IF EXISTS calendar_event_attendees;
DROP TABLE IF EXISTS calendar_events;

-- Leave management tables
DROP TABLE IF EXISTS leave_payout_requests;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS leave_balance_adjustments;
DROP TABLE IF EXISTS leave_balances;
DROP TABLE IF EXISTS leave_policies;

-- Shift swap tables
DROP TABLE IF EXISTS shift_swap_requests;

-- Late arrivals table
DROP TABLE IF EXISTS late_arrivals;

-- Payroll tables
DROP TABLE IF EXISTS pay_period_line_items;

-- Time entry table
DROP TABLE IF EXISTS time_entries;

DROP TABLE IF EXISTS pay_periods;

-- Schedule tables
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS national_holidays;

-- Handbook tables (onboarding)
DROP TABLE IF EXISTS employee_handbook_step_progress;
DROP TABLE IF EXISTS employee_handbook_assignment_history;
DROP TABLE IF EXISTS employee_handbooks;
DROP TABLE IF EXISTS handbook_steps;
DROP TABLE IF EXISTS handbook_templates;

-- Employee tables
DROP TABLE IF EXISTS employee_experience;
DROP TABLE IF EXISTS certification;
DROP TABLE IF EXISTS employee_education;
DROP TABLE IF EXISTS employee_contract_changes;
DROP TABLE IF EXISTS employee_profile CASCADE;

-- Department foreign key constraint cleanup (before dropping departments)
ALTER TABLE IF EXISTS departments DROP CONSTRAINT IF EXISTS departments_department_head_employee_id_fkey;
DROP TABLE IF EXISTS departments;

-- File management tables
DROP TABLE IF EXISTS temporary_file;
DROP TABLE IF EXISTS attachment_file;

-- Notification tables
DROP TABLE IF EXISTS notifications;

-- Session and auth tables
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_permission_overrides;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS custom_user;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

-- Location tables
DROP TABLE IF EXISTS location_shift;
DROP TABLE IF EXISTS location;

-- Organization tables
DROP TABLE IF EXISTS app_organization_profile;
DROP TABLE IF EXISTS organisations;

-- ==========================================
-- DROP ENUMS (after all tables using them)
-- ==========================================

DROP TYPE IF EXISTS reminder_channel_enum;
DROP TYPE IF EXISTS attendee_response_enum;
DROP TYPE IF EXISTS calendar_event_work_approval_status_enum;
DROP TYPE IF EXISTS calendar_event_status_enum;
DROP TYPE IF EXISTS calendar_event_kind_enum;
DROP TYPE IF EXISTS time_entry_hour_type_enum;
DROP TYPE IF EXISTS time_entry_status_enum;
DROP TYPE IF EXISTS payout_request_status_enum;
DROP TYPE IF EXISTS leave_request_status_enum;
DROP TYPE IF EXISTS leave_request_type_enum;
DROP TYPE IF EXISTS pay_period_status_enum;
DROP TYPE IF EXISTS shift_swap_status_enum;
DROP TYPE IF EXISTS handbook_assignment_event_enum;
DROP TYPE IF EXISTS handbook_step_status_enum;
DROP TYPE IF EXISTS handbook_assignment_status_enum;
DROP TYPE IF EXISTS handbook_step_kind_enum;
DROP TYPE IF EXISTS handbook_template_status_enum;
DROP TYPE IF EXISTS employee_contract_type_enum;
DROP TYPE IF EXISTS gender_enum;
DROP TYPE IF EXISTS permission_override_effect;
DROP TYPE IF EXISTS notification_type_enum;
DROP TYPE IF EXISTS location_type_enum;
DROP TYPE IF EXISTS irregular_hours_profile_enum;
