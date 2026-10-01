-- +goose Up
-- +goose StatementBegin

CREATE TYPE exec_display_group_type AS ENUM (
    'president',
    'board',
    'central_director',
    'game_director',
    'executive'
);

CREATE TYPE exec_social_platform_type AS ENUM (
    'instagram',
    'x',
    'twitch',
    'youtube',
    'tiktok',
    'linkedin'
);

CREATE TABLE exec_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    title VARCHAR NOT NULL DEFAULT 'Executive',
    display_order INT NOT NULL DEFAULT 0,
    display_group exec_display_group_type NOT NULL DEFAULT 'executive',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE exec_social_link (
    user_id UUID NOT NULL REFERENCES exec_profile(user_id) ON DELETE CASCADE,
    platform exec_social_platform_type NOT NULL,
    url TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, platform)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE exec_social_link;
DROP TABLE exec_profile;
DROP TYPE exec_social_platform_type;
DROP TYPE exec_display_group_type;
-- +gooseStatementEnd
