-- name: GetLoungeSessionsByUserId :many
SELECT
    ls.id,
    ls.game,
    ls.check_in_time,
    ls.check_out_time,
    st.label AS station_label
FROM lounge_sessions ls
JOIN lounge_stations st ON st.id = ls.station_id
WHERE ls.user_id = $1
ORDER BY ls.check_in_time DESC;
