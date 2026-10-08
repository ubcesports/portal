-- +goose Up
-- +goose StatementBegin
CREATE TYPE group_type_new AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'central_director',
    'game_director',
    'board',
    'president'
);

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'director' THEN 'game_director'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'director' THEN 'game_director'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_new
        USING (
            CASE group_at_purchase::text
                WHEN 'director' THEN 'game_director'
                ELSE group_at_purchase::text
            END
        )::group_type_new;

ALTER TABLE exec_profile
    ALTER column display_group DROP DEFAULT;

ALTER TABLE exec_profile
    ALTER COLUMN display_group TYPE group_type_new
    USING display_group::text::group_type_new;

ALTER TABLE exec_profile 
    ALTER COLUMN display_group SET DEFAULT 'executive';

DROP TYPE group_type;
DROP TYPE exec_display_group_type;

ALTER TYPE group_type_new RENAME TO group_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TYPE group_type_old AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'director',
    'board'
);

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
    USING display_group::text::exec_display_group_type;

ALTER TABLE exec_profile
    ALTER COLUMN display_group SET DEFAULT 'executive';

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_old
    USING (
        CASE "group"::text
            WHEN 'central_director' THEN 'director'
            WHEN 'game_director' THEN 'director'
            WHEN 'president' THEN 'board'
            ELSE "group"::text
        END
    )::group_type_old;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_old
    USING (
        CASE "group"::text
            WHEN 'central_director' THEN 'director'
            WHEN 'game_director' THEN 'director'
            WHEN 'president' THEN 'board'
            ELSE "group"::text
        END
    )::group_type_old;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_old
    USING (
        CASE group_at_purchase::text
            WHEN 'central_director' THEN 'director'
            WHEN 'game_director' THEN 'director'
            WHEN 'president' THEN 'board'
            ELSE group_at_purchase::text
        END
    )::group_type_old;

DROP TYPE group_type;
ALTER TYPE group_type_old RENAME TO group_type;

-- +goose StatementEnd
