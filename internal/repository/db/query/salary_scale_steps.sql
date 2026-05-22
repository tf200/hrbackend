-- name: ListSalaryScaleSteps :many
SELECT
    css.id,
    css.salary_table_id,
    css.scale,
    css.step,
    css.ip_number,
    css.monthly_salary,
    css.hourly_rate,
    cst.cao_code,
    cst.name AS salary_table_name,
    cst.effective_from,
    cst.effective_to
FROM cao_salary_scale_steps css
JOIN cao_salary_tables cst ON cst.id = css.salary_table_id
WHERE (CASE WHEN sqlc.narg('active_only')::bool THEN cst.effective_from <= CURRENT_DATE AND (cst.effective_to IS NULL OR cst.effective_to >= CURRENT_DATE) ELSE TRUE END)
ORDER BY cst.effective_from DESC, css.scale ASC, css.step ASC;
