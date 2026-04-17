-- name: GetAppOrganizationProfile :one
SELECT
    name,
    default_timezone,
    email,
    phone_number,
    website,
    hq_street,
    hq_house_number,
    hq_house_number_addition,
    hq_postal_code,
    hq_city,
    created_at,
    updated_at
FROM app_organization_profile
WHERE singleton = TRUE;

-- name: UpdateAppOrganizationProfile :one
UPDATE app_organization_profile
SET
    name = COALESCE(sqlc.narg('name'), name),
    default_timezone = COALESCE(sqlc.narg('default_timezone'), default_timezone),
    email = COALESCE(sqlc.narg('email'), email),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    website = COALESCE(sqlc.narg('website'), website),
    hq_street = COALESCE(sqlc.narg('hq_street'), hq_street),
    hq_house_number = COALESCE(sqlc.narg('hq_house_number'), hq_house_number),
    hq_house_number_addition = COALESCE(
        sqlc.narg('hq_house_number_addition'),
        hq_house_number_addition
    ),
    hq_postal_code = COALESCE(sqlc.narg('hq_postal_code'), hq_postal_code),
    hq_city = COALESCE(sqlc.narg('hq_city'), hq_city),
    updated_at = NOW()
WHERE singleton = TRUE
RETURNING
    name,
    default_timezone,
    email,
    phone_number,
    website,
    hq_street,
    hq_house_number,
    hq_house_number_addition,
    hq_postal_code,
    hq_city,
    created_at,
    updated_at;
