-- name: GetActiveMembershipsCount :one
WITH args AS (
    SELECT
        sqlc.narg('program_name')::text AS program_name,
        sqlc.narg('tier_ids')::text[] AS tier_ids,
        sqlc.narg('is_student')::boolean AS is_student
)
SELECT COUNT(*)
FROM memberships m
JOIN membership_tiers mt ON mt.id = m.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
JOIN users u ON u.id = m.user_id
CROSS JOIN args a
WHERE m.cancelled_at IS NULL
  AND m.started_at <= NOW()
  AND m.expires_at > NOW()
  AND (a.program_name IS NULL OR mp.program_name = a.program_name)
  AND (a.tier_ids IS NULL OR mt.id::text = ANY(a.tier_ids))
  AND (a.is_student IS NULL OR u.is_student = a.is_student);

-- name: GetAllTimeRevenueCents :one
WITH args AS (
    SELECT
        sqlc.narg('program_name')::text AS program_name,
        sqlc.narg('tier_ids')::text[] AS tier_ids,
        sqlc.narg('is_student')::boolean AS is_student,
        sqlc.narg('purchase_type')::purchase_type AS purchase_type
)
SELECT COALESCE(SUM(t.amount_paid_cents), 0)::bigint AS revenue_cents
FROM transactions t
JOIN membership_tiers mt ON mt.id = t.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
CROSS JOIN args a
WHERE t.status = 'completed'
  AND (a.program_name IS NULL OR mp.program_name = a.program_name)
  AND (a.tier_ids IS NULL OR mt.id::text = ANY(a.tier_ids))
  AND (a.is_student IS NULL OR t.student_at_purchase = a.is_student)
  AND (a.purchase_type IS NULL OR t.purchase_type = a.purchase_type);

-- name: GetUniqueMembersCount :one
WITH args AS (
    SELECT
        sqlc.narg('program_name')::text AS program_name,
        sqlc.narg('tier_ids')::text[] AS tier_ids,
        sqlc.narg('is_student')::boolean AS is_student,
        sqlc.narg('purchase_type')::purchase_type AS purchase_type
)
SELECT COUNT(DISTINCT t.user_id)
FROM transactions t
JOIN membership_tiers mt ON mt.id = t.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
CROSS JOIN args a
WHERE t.status = 'completed'
  AND (a.program_name IS NULL OR mp.program_name = a.program_name)
  AND (a.tier_ids IS NULL OR mt.id::text = ANY(a.tier_ids))
  AND (a.is_student IS NULL OR t.student_at_purchase = a.is_student)
  AND (a.purchase_type IS NULL OR t.purchase_type = a.purchase_type);

-- name: GetMembershipsBoughtOverTime :many
WITH args AS (
    SELECT
        sqlc.arg('granularity')::text AS granularity,
        sqlc.narg('from_date')::timestamptz AS from_date,
        sqlc.arg('to_date')::timestamptz AS to_date,
        sqlc.narg('program_name')::text AS program_name,
        sqlc.narg('tier_ids')::text[] AS tier_ids,
        sqlc.narg('is_student')::boolean AS is_student
)
SELECT
    date_trunc((SELECT granularity FROM args), t.created_at)::timestamptz AS bucket,
    COUNT(*) AS count
FROM transactions t
JOIN membership_tiers mt ON mt.id = t.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
CROSS JOIN args a
WHERE t.status = 'completed'
  AND (a.from_date IS NULL OR t.created_at >= a.from_date)
  AND t.created_at < a.to_date
  AND (a.program_name IS NULL OR mp.program_name = a.program_name)
  AND (a.tier_ids IS NULL OR mt.id::text = ANY(a.tier_ids))
  AND (a.is_student IS NULL OR t.student_at_purchase = a.is_student)
GROUP BY bucket
ORDER BY bucket;

-- name: GetRevenueOverTime :many
WITH args AS (
    SELECT
        sqlc.arg('granularity')::text AS granularity,
        sqlc.narg('from_date')::timestamptz AS from_date,
        sqlc.arg('to_date')::timestamptz AS to_date,
        sqlc.narg('program_name')::text AS program_name,
        sqlc.narg('tier_ids')::text[] AS tier_ids,
        sqlc.narg('is_student')::boolean AS is_student,
        sqlc.narg('purchase_type')::purchase_type AS purchase_type
)
SELECT
    date_trunc((SELECT granularity FROM args), t.created_at)::timestamptz AS bucket,
    COALESCE(SUM(t.amount_paid_cents), 0)::bigint AS revenue_cents
FROM transactions t
JOIN membership_tiers mt ON mt.id = t.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
CROSS JOIN args a
WHERE t.status = 'completed'
  AND (a.from_date IS NULL OR t.created_at >= a.from_date)
  AND t.created_at < a.to_date
  AND (a.program_name IS NULL OR mp.program_name = a.program_name)
  AND (a.tier_ids IS NULL OR mt.id::text = ANY(a.tier_ids))
  AND (a.is_student IS NULL OR t.student_at_purchase = a.is_student)
  AND (a.purchase_type IS NULL OR t.purchase_type = a.purchase_type)
GROUP BY bucket
ORDER BY bucket;
