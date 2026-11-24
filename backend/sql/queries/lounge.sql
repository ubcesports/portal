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

-- name: CheckoutLoungeSession :one
WITH updated_session AS (
    UPDATE lounge_sessions
    SET 
        check_out_time = NOW(),
        checked_out_by = sqlc.narg('checked_out_by')
    WHERE id = sqlc.arg('id') AND check_out_time IS NULL
    RETURNING 
        id,
        user_id,
        membership_id,
        station_id,
        check_in_time,
        check_out_time,
        checked_in_by,
        checked_out_by,
        game
)
SELECT 
    s.id,
    s.user_id,
    s.membership_id,
    s.station_id,
    s.check_in_time,
    COALESCE(u.check_out_time, s.check_out_time) AS check_out_time,
    s.checked_in_by,
    COALESCE(u.checked_out_by, s.checked_out_by) AS checked_out_by,
    s.game,
    (u.id IS NULL)::boolean AS already_checked_out
FROM lounge_sessions s
LEFT JOIN updated_session u ON u.id = s.id
WHERE s.id = sqlc.arg('id');
