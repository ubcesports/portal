-- +goose Up
-- +goose StatementBegin

ALTER TABLE membership_tiers
ADD COLUMN lounge_access BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE membership_tiers
SET lounge_access = TRUE
WHERE slug IN ('day', 'lounge', 'executive', 'competitive_team');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE membership_tiers
DROP COLUMN lounge_access;
-- +goose StatementEnd
