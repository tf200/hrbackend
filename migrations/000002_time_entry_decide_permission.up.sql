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
VALUES (
    'TIME_ENTRY.DECIDE',
    'TIME_ENTRY',
    'DECIDE',
    'time_entry',
    'decide',
    'Time Entry Decide',
    NULL,
    495
)
ON CONFLICT (name) DO UPDATE SET
    resource = EXCLUDED.resource,
    method = EXCLUDED.method,
    group_key = EXCLUDED.group_key,
    section_key = EXCLUDED.section_key,
    sort_order = EXCLUDED.sort_order;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name = 'TIME_ENTRY.DECIDE'
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;
