-- +goose Up
-- +goose StatementBegin
CREATE TABLE lounge_stations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label VARCHAR NOT NULL,
    station_type VARCHAR NOT NULL,
    closed BOOLEAN NOT NULL DEFAULT FALSE,
    "row" INT NOT NULL,
    col INT NOT NULL,
    width INT NOT NULL DEFAULT 1,
    height INT NOT NULL DEFAULT 1,
    active_session_id UUID
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lounge_stations;
-- +goose StatementEnd
