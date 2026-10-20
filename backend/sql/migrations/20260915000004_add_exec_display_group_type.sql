-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'exec_profile'
          AND column_name = 'display_group'
    ) THEN
        ALTER TABLE exec_profile
            ADD COLUMN display_group group_type NOT NULL DEFAULT 'executive';
    END IF;
END $$;

CREATE TYPE exec_display_group_type AS ENUM (
    'president',
    'board',
    'central_director',
    'game_director',
    'executive'
);

ALTER TABLE exec_profile
    ALTER COLUMN display_group DROP DEFAULT;

ALTER TABLE exec_profile
    ALTER COLUMN display_group TYPE exec_display_group_type
        USING (
            CASE display_group::text
                WHEN 'president' THEN 'president'::exec_display_group_type
                WHEN 'board' THEN 'board'::exec_display_group_type
                WHEN 'central_director' THEN 'central_director'::exec_display_group_type
                WHEN 'game_director' THEN 'game_director'::exec_display_group_type
                ELSE 'executive'::exec_display_group_type
            END
        );

ALTER TABLE exec_profile
    ALTER COLUMN display_group SET DEFAULT 'executive';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE exec_profile
    ALTER COLUMN display_group DROP DEFAULT;

ALTER TABLE exec_profile
    ALTER COLUMN display_group TYPE group_type
        USING (
            CASE display_group::text
                WHEN 'president' THEN 'president'::group_type
                WHEN 'board' THEN 'board'::group_type
                WHEN 'central_director' THEN 'director'::group_type
                WHEN 'game_director' THEN 'director'::group_type
                ELSE 'executive'::group_type
            END
        );

ALTER TABLE exec_profile
    ALTER COLUMN display_group SET DEFAULT 'executive';

DROP TYPE IF EXISTS exec_display_group_type;
-- +goose StatementEnd