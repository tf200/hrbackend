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

WITH seeded(name, sort_order) AS (
    VALUES
        ('PERFORMANCE.ASSESSMENT.CREATE', 560),
        ('PERFORMANCE.ASSESSMENT.VIEW', 570),
        ('PERFORMANCE.ASSESSMENT.VIEW_ALL', 580),
        ('PERFORMANCE.ASSESSMENT.DELETE', 590),
        ('PERFORMANCE.WORK_ASSIGNMENT.VIEW', 600),
        ('PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL', 610),
        ('PERFORMANCE.WORK_ASSIGNMENT.DECIDE', 620),
        ('PERFORMANCE.UPCOMING.INVITE', 630),
        ('PERFORMANCE.STATS.VIEW', 640)
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

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT ar.id, p.id
FROM admin_role ar
JOIN permissions p ON p.name IN (
    'PERFORMANCE.ASSESSMENT.CREATE',
    'PERFORMANCE.ASSESSMENT.VIEW',
    'PERFORMANCE.ASSESSMENT.VIEW_ALL',
    'PERFORMANCE.ASSESSMENT.DELETE',
    'PERFORMANCE.WORK_ASSIGNMENT.VIEW',
    'PERFORMANCE.WORK_ASSIGNMENT.VIEW_ALL',
    'PERFORMANCE.WORK_ASSIGNMENT.DECIDE',
    'PERFORMANCE.UPCOMING.INVITE',
    'PERFORMANCE.STATS.VIEW'
)
ON CONFLICT DO NOTHING;
