-- +goose Up
-- +goose StatementBegin
CREATE TYPE group_type_new AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'director',
    'board',
    'president'
);

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'central_director' THEN 'director'
                WHEN 'game_director' THEN 'director'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'central_director' THEN 'director'
                WHEN 'game_director' THEN 'director'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_new
        USING (
            CASE group_at_purchase::text
                WHEN 'central_director' THEN 'director'
                WHEN 'game_director' THEN 'director'
                ELSE group_at_purchase::text
            END
        )::group_type_new;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'exec_profile'
          AND column_name = 'display_group'
    ) THEN
        ALTER TABLE exec_profile
            ALTER COLUMN display_group TYPE TEXT
                USING display_group::text;
    END IF;
END $$;

DROP TYPE IF EXISTS group_type CASCADE;
ALTER TYPE group_type_new RENAME TO group_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TYPE group_type_old AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'central_director',
    'game_director',
    'board',
    'president'
);

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_old
        USING (
            CASE "group"::text
                WHEN 'director' THEN 'game_director'
                ELSE "group"::text
            END
        )::group_type_old;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_old
        USING (
            CASE "group"::text
                WHEN 'director' THEN 'game_director'
                ELSE "group"::text
            END
        )::group_type_old;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_old
        USING (
            CASE group_at_purchase::text
                WHEN 'director' THEN 'game_director'
                ELSE group_at_purchase::text
            END
        )::group_type_old;

DROP TYPE IF EXISTS group_type CASCADE;
ALTER TYPE group_type_old RENAME TO group_type;
-- +goose StatementEnd
