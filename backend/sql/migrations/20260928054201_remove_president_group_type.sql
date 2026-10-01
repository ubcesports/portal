-- +goose Up
-- +goose StatementBegin
CREATE TYPE group_type_new AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'director',
    'board'
);

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'president' THEN 'board'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_new
        USING (
            CASE "group"::text
                WHEN 'president' THEN 'board'
                ELSE "group"::text
            END
        )::group_type_new;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_new
        USING (
            CASE group_at_purchase::text
                WHEN 'president' THEN 'board'
                ELSE group_at_purchase::text
            END
        )::group_type_new;

DROP TYPE IF EXISTS group_type CASCADE;
ALTER TYPE group_type_new RENAME TO group_type;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TYPE group_type_old AS ENUM (
    'member',
    'competitive_team',
    'executive',
    'director',
    'board',
    'president'
);

ALTER TABLE user_groups
    ALTER COLUMN "group" TYPE group_type_old
        USING "group"::text::group_type_old;

ALTER TABLE membership_tiers
    ALTER COLUMN "group" TYPE group_type_old
        USING "group"::text::group_type_old;

ALTER TABLE transactions
    ALTER COLUMN group_at_purchase TYPE group_type_old
        USING group_at_purchase::text::group_type_old;

DROP TYPE IF EXISTS group_type CASCADE;
ALTER TYPE group_type_old RENAME TO group_type;
-- +goose StatementEnd
