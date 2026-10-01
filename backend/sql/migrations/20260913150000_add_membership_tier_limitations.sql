-- +goose Up
-- +goose StatementBegin
ALTER TABLE membership_tiers
    ADD COLUMN limitations TEXT[] NOT NULL DEFAULT '{}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE membership_tiers
    DROP COLUMN IF EXISTS limitations;
-- +goose StatementEnd
