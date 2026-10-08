-- name: GetProfileByUserID :one
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
    SELECT array_agg(ug."group"::text ORDER BY ug.assigned_at ASC, ug."group" ASC) AS groups
    FROM user_groups ug
    WHERE ug.user_id = u.id
) g ON true
WHERE u.id = $1
;

-- name: EnsureMemberGroupForUser :exec
INSERT INTO user_groups (
    user_id,
    "group"
)
VALUES (
    $1,
    'member'
)
ON CONFLICT (user_id, "group") DO NOTHING;

-- name: OnboardUserByUserId :exec
WITH onboarded AS (
    UPDATE users
    SET
        is_student = sqlc.arg(is_student),
        student_id = sqlc.arg(student_id),
        full_name = sqlc.arg(full_name),
        onboarding_completed_at = NOW(),
        updated_at = NOW()
    WHERE id = sqlc.arg(id)
      AND onboarding_completed_at IS NULL
    RETURNING id
), executive_group AS (
    INSERT INTO user_groups (user_id, "group")
    SELECT id, 'executive'::group_type
    FROM onboarded
    WHERE sqlc.arg(is_executive)::boolean
    ON CONFLICT (user_id, "group") DO NOTHING
)
INSERT INTO exec_profile (user_id)
SELECT id
FROM onboarded
WHERE sqlc.arg(is_executive)::boolean
ON CONFLICT (user_id) DO NOTHING;
