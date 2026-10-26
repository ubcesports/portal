-- +goose Up
ALTER TABLE users
ADD COLUMN admin_note TEXT DEFAULT '';

-- +goose Down
ALTER TABLE users
DROP COLUMN admin_note;
