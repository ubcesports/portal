-- +goose Up
-- +goose StatementBegin

INSERT INTO exec_profile(
    user_id,
    title,
    display_order,
    display_group
)
SELECT
    u.id,
    'Executive',
    0,
    CASE
        WHEN EXISTS (
            SELECT 1
            FROM user_groups ug
            WHERE ug.user_id = u.id
                AND ug."group" = 'board'
        ) THEN 'board'::exec_display_group_type
        WHEN EXISTS (
            SELECT 1
            FROM user_groups ug
            WHERE ug.user_id = u.id
                AND ug."group" = 'director'
        ) THEN 'game_director'::exec_display_group_type
        ELSE 'executive'::exec_display_group_type
    END
FROM users u
WHERE EXISTS (
    SELECT 1
    FROM user_groups ug
    WHERE ug.user_id = u.id 
        AND ug."group" IN ('executive', 'director', 'board')
)
ON CONFLICT (user_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
-- be careful with this, as it will delete ALL currently 
-- existing exec profiles matching this condition
DELETE FROM exec_profile
WHERE user_id IN (
    SELECT ug.user_id
    FROM user_groups ug
    WHERE ug."group" IN ('executive', 'director', 'board')
)
-- +goose StatementEnd
