-- +goose Up
-- +goose StatementBegin

ALTER TABLE membership_tiers
ADD COLUMN lounge_access BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE membership_tiers
DROP COLUMN lounge_access;
-- +goose StatementEnd
