-- name: GetUsersAdmin :many
WITH args AS (
    SELECT
        sqlc.narg('full_name')::text AS full_name,
        sqlc.narg('student_id')::text AS student_id,
        sqlc.narg('email')::text AS email,
        sqlc.narg('role')::role_type AS role,
        sqlc.narg('is_student')::boolean AS is_student,
        sqlc.narg('groups')::text[] AS groups,
        sqlc.narg('membership_tier_ids')::text[] AS membership_tier_ids,
        sqlc.narg('limit')::integer AS "limit",
        sqlc.narg('offset')::integer AS "offset"
)
SELECT
    u.id,
    u.email,
    u.student_id,
    u.role,
    u.created_at,
    u.updated_at,
    u.full_name,
    u.email_verified_at,
    u.is_student,
    u.onboarding_completed_at,
    u.avatar_url,
    COALESCE(g.groups, '{}'::text[])::text[] AS groups,
    COALESCE(am.tier_titles, '{}'::text[])::text[] AS active_membership_tier_titles
FROM users u
CROSS JOIN args a
LEFT JOIN LATERAL (
    SELECT array_agg(
        ug."group"::text
        ORDER BY ug.assigned_at, ug."group"
    ) AS groups
    FROM user_groups ug
    WHERE ug.user_id = u.id
) g ON true
LEFT JOIN LATERAL (
    SELECT
        array_agg(mt.title ORDER BY mt.title, mt.id) AS tier_titles
    FROM memberships m
    JOIN membership_tiers mt ON mt.id = m.tier_id
    WHERE m.user_id = u.id
      AND m.cancelled_at IS NULL
      AND m.started_at <= NOW()
      AND m.expires_at > NOW()
) am ON true
WHERE (
    a.full_name IS NULL
    OR u.full_name ILIKE '%' || a.full_name || '%'
)
AND (
    a.student_id IS NULL
    OR u.student_id ILIKE '%' || a.student_id || '%'
)
AND (
    a.email IS NULL
    OR u.email ILIKE '%' || a.email || '%'
)
AND (
    a.role IS NULL
    OR u.role = a.role
)
AND (
    a.is_student IS NULL
    OR u.is_student = a.is_student
)
AND (
    a.groups IS NULL
    OR ARRAY(
        SELECT filter_group."group"::text
        FROM user_groups filter_group
        WHERE filter_group.user_id = u.id
        ORDER BY filter_group."group"::text
    ) = a.groups
)
AND (
    a.membership_tier_ids IS NULL
    OR ARRAY(
        SELECT DISTINCT filter_tier.tier_id::text
        FROM memberships filter_tier
        WHERE filter_tier.user_id = u.id
          AND filter_tier.cancelled_at IS NULL
          AND filter_tier.started_at <= NOW()
          AND filter_tier.expires_at > NOW()
        ORDER BY filter_tier.tier_id::text
    ) = a.membership_tier_ids
)
ORDER BY u.created_at DESC
LIMIT (SELECT "limit" FROM args)
OFFSET (SELECT "offset" FROM args);

-- name: CountUsersAdmin :one
WITH args AS (
    SELECT
        sqlc.narg('full_name')::text AS full_name,
        sqlc.narg('student_id')::text AS student_id,
        sqlc.narg('email')::text AS email,
        sqlc.narg('role')::role_type AS role,
        sqlc.narg('is_student')::boolean AS is_student,
        sqlc.narg('groups')::text[] AS groups,
        sqlc.narg('membership_tier_ids')::text[] AS membership_tier_ids
)
SELECT COUNT(*)
FROM users u
CROSS JOIN args a
WHERE (
    a.full_name IS NULL
    OR u.full_name ILIKE '%' || a.full_name || '%'
)
AND (
    a.student_id IS NULL
    OR u.student_id ILIKE '%' || a.student_id || '%'
)
AND (
    a.email IS NULL
    OR u.email ILIKE '%' || a.email || '%'
)
AND (
    a.role IS NULL
    OR u.role = a.role
)
AND (
    a.is_student IS NULL
    OR u.is_student = a.is_student
)
AND (
    a.groups IS NULL
    OR ARRAY(
        SELECT filter_group."group"::text
        FROM user_groups filter_group
        WHERE filter_group.user_id = u.id
        ORDER BY filter_group."group"::text
    ) = a.groups
)
AND (
    a.membership_tier_ids IS NULL
    OR ARRAY(
        SELECT DISTINCT filter_tier.tier_id::text
        FROM memberships filter_tier
        WHERE filter_tier.user_id = u.id
          AND filter_tier.cancelled_at IS NULL
          AND filter_tier.started_at <= NOW()
          AND filter_tier.expires_at > NOW()
        ORDER BY filter_tier.tier_id::text
    ) = a.membership_tier_ids
);

-- name: GetAdminMembershipTierOptions :many
SELECT
    mt.id,
    mt.title,
    mp.program_name
FROM membership_tiers mt
JOIN membership_programs mp ON mp.id = mt.program_id
ORDER BY mp.program_name, mt.title, mt.id;

-- name: GetAdminUserByID :one
SELECT
    u.id,
    u.email,
    u.student_id,
    u.role,
    u.created_at,
    u.updated_at,
    u.full_name,
    u.email_verified_at,
    u.is_student,
    u.onboarding_completed_at,
    u.avatar_url,
    COALESCE(g.groups, '{}'::text[])::text[] AS groups
FROM users u
LEFT JOIN LATERAL (
    SELECT array_agg(
        ug."group"::text
        ORDER BY ug.assigned_at, ug."group"
    ) AS groups
    FROM user_groups ug
    WHERE ug.user_id = u.id
) g ON true
WHERE u.id = $1;

-- name: UpdateUserFullName :exec
UPDATE users
SET full_name = sqlc.arg(full_name), updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: UpdateUserStudentInfo :exec
UPDATE users
SET
    is_student = $2,
    student_id = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: StudentIDExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE student_id = $1
);

-- name: UpdateUserRole :exec
UPDATE users
SET
    role = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: AddUserGroup :exec
INSERT INTO user_groups (user_id, "group")
VALUES ($1, $2)
ON CONFLICT (user_id, "group") DO NOTHING;

-- name: RemoveUserGroup :exec
DELETE FROM user_groups
WHERE user_id = $1 AND "group" = $2;

-- name: HasActiveMembershipForUser :one
SELECT EXISTS (
    SELECT 1
    FROM memberships
    WHERE user_id = $1
      AND cancelled_at IS NULL
);

-- name: CreateAdminAuditLog :exec
INSERT INTO admin_audit_logs (
    actor_user_id,
    action,
    target_user_id,
    outcome,
    request_id,
    description
) VALUES (
    sqlc.arg('actor_user_id')::uuid,
    sqlc.arg('action')::text,
    sqlc.narg('target_user_id')::uuid,
    sqlc.arg('outcome')::admin_audit_outcome_type,
    sqlc.arg('request_id')::text,
    sqlc.narg('description')::text
);

-- name: GetAdminAuditLogs :many
WITH args AS (
    SELECT
        sqlc.narg('actor_name')::text AS actor_name,
        sqlc.arg('limit')::integer AS "limit",
        sqlc.arg('offset')::integer AS "offset"
)
SELECT
    aal.id,
    aal.occurred_at,
    aal.action,
    aal.outcome,
    aal.request_id,
    aal.description,
    actor.id AS actor_id,
    actor.full_name AS actor_name,
    actor.avatar_url AS actor_avatar_url,
    target.id AS target_id,
    target.full_name AS target_name,
    target.avatar_url AS target_avatar_url
FROM admin_audit_logs aal
JOIN users actor
    ON actor.id = aal.actor_user_id
LEFT JOIN users target
    ON target.id = aal.target_user_id
CROSS JOIN args a
WHERE (
    a.actor_name IS NULL
    OR actor.full_name ILIKE '%' || a.actor_name || '%'
)
ORDER BY
    aal.occurred_at DESC,
    aal.id DESC
LIMIT (SELECT "limit" FROM args)
OFFSET (SELECT "offset" FROM args);

-- name: CountAdminAuditLogs :one
SELECT COUNT(*)
FROM admin_audit_logs aal
JOIN users actor
    ON actor.id = aal.actor_user_id
WHERE (
    sqlc.narg(actor_name)::text IS NULL
    OR actor.full_name ILIKE '%' || sqlc.narg(actor_name)::text || '%'
);


-- name: CreateExecProfile :exec
INSERT INTO exec_profile (
    user_id, 
    title,
    display_order,
    display_group
) VALUES (
    sqlc.narg('user_id')::uuid,
    sqlc.narg('title')::text,
    sqlc.narg('display_order')::int,
    sqlc.narg('display_group')::exec_display_group_type
);

-- name: HasExecProfileForUser :one
SELECT EXISTS (
    SELECT 1 
    FROM exec_profile
    WHERE user_id = $1
);

-- name: HasExecGroupForUser :one
SELECT EXISTS (
    SELECT 1 
    FROM user_groups
    WHERE user_id = $1
        AND "group" IN ('executive', 'director', 'board')
);

-- name: RemoveExecProfile :exec
DELETE FROM exec_profile
WHERE user_id = $1;

-- name: UpdateExecProfile :exec
UPDATE exec_profile
SET
    title = COALESCE(sqlc.narg('title'), title),
    display_order = COALESCE(sqlc.narg('display_order'), display_order),
    display_group = COALESCE(sqlc.narg('display_group'), display_group),
    updated_at = NOW()
WHERE user_id = sqlc.narg('user_id');

-- name: GetExecProfileByUserID :one
SELECT
    user_id,
    title,
    display_order,
    display_group
FROM exec_profile
WHERE user_id = $1;