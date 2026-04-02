DELETE FROM role_permissions rp
USING roles r, permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.name = 'admin'
  AND p.name = 'TIME_ENTRY.DECIDE';

DELETE FROM permissions
WHERE name = 'TIME_ENTRY.DECIDE';
