-- name: GetRoles :many
SELECT
    r.id,
    r.name,
    r.description,
    r.color,
    r.display_order,
    r.created_by_user_id,
    creator.full_name AS created_by_name,
    creator.avatar_url AS created_by_profile_picture,
    r.created_at,
    r.updated_at
FROM roles r
LEFT JOIN users creator ON creator.id = r.created_by_user_id
ORDER BY r.display_order, r.name;

-- name: GetRoleByID :one
SELECT
    r.id,
    r.name,
    r.description,
    r.color,
    r.display_order,
    r.created_by_user_id,
    creator.full_name AS created_by_name,
    creator.avatar_url AS created_by_profile_picture,
    ARRAY(
        SELECT rp.permission
        FROM role_permissions rp
        WHERE rp.role_id = r.id
        ORDER BY rp.permission
    )::TEXT[] AS permission_keys,
    COALESCE(
        (
            SELECT json_agg(
                json_build_object(
                    'user_id', u.id,
                    'name', u.full_name,
                    'profile_picture', u.avatar_url
                )
                ORDER BY u.full_name, u.id
            )
            FROM user_roles ur
            JOIN users u ON u.id = ur.user_id
            WHERE ur.role_id = r.id
        ),
        '[]'::JSON
    ) AS users_with_role,
    r.created_at,
    r.updated_at
FROM roles r
LEFT JOIN users creator ON creator.id = r.created_by_user_id
WHERE r.id = $1;

-- name: CreateRole :one
INSERT INTO roles (
    name,
    description,
    color,
    display_order,
    created_by_user_id
)
VALUES (
    sqlc.arg('name'),
    sqlc.narg('description'),
    sqlc.narg('color'),
    sqlc.arg('display_order'),
    sqlc.arg('created_by_user_id')
)
RETURNING id;

-- name: UpdateRole :one
UPDATE roles
SET
    name = COALESCE(sqlc.narg('name')::VARCHAR(50), name),
    description = COALESCE(sqlc.narg('description')::TEXT, description),
    color = COALESCE(sqlc.narg('color')::VARCHAR(7), color),
    display_order = COALESCE(sqlc.narg('display_order')::INTEGER, display_order),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING
    id,
    name,
    description,
    color,
    display_order,
    created_by_user_id,
    created_at,
    updated_at;

-- name: AddRolePermissionKeys :many
INSERT INTO role_permissions (role_id, permission)
SELECT sqlc.arg('role_id'), permission
FROM UNNEST(sqlc.arg('permission_keys')::TEXT[]) AS permission
ON CONFLICT (role_id, permission) DO NOTHING
RETURNING permission;

-- name: DeleteRolePermissionKeys :exec
DELETE FROM role_permissions
WHERE role_id = $1;

-- name: DeleteRole :execrows
WITH deleted_user_roles AS (
    DELETE FROM user_roles
    WHERE role_id = $1
    RETURNING role_id
),
deleted_role_permissions AS (
    DELETE FROM role_permissions
    WHERE role_id = $1
    RETURNING role_id
)
DELETE FROM roles
WHERE id = $1;

-- name: GetRolesForUser :many
SELECT
    r.id,
    r.name,
    rp.permission
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
LEFT JOIN role_permissions rp ON rp.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.display_order, r.name, rp.permission;

-- name: UserHasAnyRole :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE ur.user_id = sqlc.arg('user_id')
      AND r.name = ANY(sqlc.arg('role_names')::TEXT[])
);

-- name: GetPermissionsForUser :many
SELECT DISTINCT rp.permission
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
WHERE ur.user_id = $1
ORDER BY rp.permission;

-- name: GetRoleAssignee :one
SELECT id, full_name
FROM users
WHERE id = $1;

-- name: AssignUserRole :execrows
INSERT INTO user_roles (user_id, role_id, assigned_by_user_id)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('role_id'),
    sqlc.arg('assigned_by_user_id')
)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: UnassignUserRole :execrows
DELETE FROM user_roles
WHERE user_id = $1
  AND role_id = $2;
