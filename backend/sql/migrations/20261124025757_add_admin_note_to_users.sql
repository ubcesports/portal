-- +goose Up
ALTER TABLE users
ADD COLUMN admin_note TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users
DROP COLUMN admin_note;
