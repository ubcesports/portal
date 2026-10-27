-- name: CreateLoungeSession :one
INSERT INTO lounge_sessions (
    user_id,
    membership_id,
    station_id,
    checked_in_by,
    game
)
SELECT
    u.id,
    sqlc.arg(membership_id),
    sqlc.arg(station_id),
    sqlc.arg(checked_in_by),
    sqlc.narg(game)
FROM users AS u
WHERE u.student_id = sqlc.arg(student_id)
RETURNING
    id,
    user_id,
    membership_id,
    station_id,
    check_in_time,
    checked_in_by,
    game;

-- name: GetLoungeEligibleMembershipByStudentID :one
SELECT
    u.id AS user_id,
    m.id AS membership_id
FROM users AS u
JOIN memberships AS m
    ON m.user_id = u.id
JOIN membership_tiers AS mt
    ON mt.id = m.tier_id
WHERE u.student_id = sqlc.arg(student_id)
  AND m.cancelled_at IS NULL
  AND m.started_at <= NOW()
  AND m.expires_at > NOW()
  AND mt.lounge_access IS TRUE
ORDER BY
    (mt.slug = 'lounge') DESC,
    m.started_at DESC,
    m.created_at DESC
LIMIT 1;

-- name: GetActiveLoungeSessionByUserID :one
SELECT id
FROM lounge_sessions
WHERE user_id = sqlc.arg(user_id)
  AND check_out_time IS NULL
LIMIT 1;

-- name: GetLoungeStations :many
SELECT 
    ls.id,
    ls.label,
    ls.station_type,
    ls.closed,
    ls.grid_row,
    ls.grid_column,
    ls.width,
    ls.height,
    ls.created_at,
    ls.updated_at,
    s.id AS active_session_id
FROM lounge_stations ls
LEFT JOIN lounge_sessions s 
    ON ls.id = s.station_id AND s.check_out_time IS NULL
ORDER BY ls.grid_row ASC, ls.grid_column ASC, ls.label ASC;

-- name: GetLoungeStationByID :one
SELECT 
    id,
    label,
    station_type,
    closed,
    grid_row,
    grid_column,
    width,
    height,
    created_at,
    updated_at
FROM lounge_stations
WHERE id = $1;

-- name: GetLoungeStationByIDWithActiveSession :one
SELECT 
    ls.id,
    ls.label,
    ls.station_type,
    ls.closed,
    ls.grid_row,
    ls.grid_column,
    ls.width,
    ls.height,
    ls.created_at,
    ls.updated_at,
    s.id AS active_session_id
FROM lounge_stations ls
LEFT JOIN lounge_sessions s 
    ON ls.id = s.station_id AND s.check_out_time IS NULL
WHERE ls.id = $1;

-- name: UpdateLoungeStation :one
UPDATE lounge_stations
SET 
    label = COALESCE(sqlc.narg('label'), label),
    closed = COALESCE(sqlc.narg('closed'), closed),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING 
    id,
    label,
    station_type,
    closed,
    grid_row,
    grid_column,
    width,
    height,
    created_at,
    updated_at;
