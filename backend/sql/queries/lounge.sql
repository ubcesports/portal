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
