-- +goose Up
-- +goose StatementBegin
CREATE TYPE lounge_station_type AS ENUM (
  'pc',
  'cab',
  'check_in'
);

CREATE TABLE lounge_stations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code VARCHAR NOT NULL UNIQUE,
  label VARCHAR NOT NULL,
  station_type lounge_station_type NOT NULL,
  closed BOOLEAN NOT NULL DEFAULT false,
  grid_row SMALLINT NOT NULL,
  grid_column SMALLINT NOT NULL,
  width SMALLINT NOT NULL DEFAULT 1,
  height SMALLINT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CHECK (grid_row >= 0),
  CHECK (grid_column >= 0),
  CHECK (width > 0),
  CHECK (height > 0)
);

CREATE TABLE lounge_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  membership_id UUID REFERENCES memberships(id),
  station_id UUID NOT NULL REFERENCES lounge_stations(id),
  check_in_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  check_out_time TIMESTAMPTZ,
  checked_in_by UUID NOT NULL REFERENCES users(id),
  checked_out_by UUID REFERENCES users(id),
  game VARCHAR,

  CHECK (
    check_out_time IS NULL
    OR check_out_time >= check_in_time
  )
);

CREATE UNIQUE INDEX lounge_sessions_one_active_per_station
  ON lounge_sessions (station_id)
  WHERE check_out_time IS NULL;

CREATE UNIQUE INDEX lounge_sessions_one_active_per_user
  ON lounge_sessions (user_id)
  WHERE check_out_time IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS lounge_sessions_one_active_per_station;
DROP INDEX IF EXISTS lounge_sessions_one_active_per_user;

DROP TABLE IF EXISTS lounge_sessions;
DROP TABLE IF EXISTS lounge_stations;

DROP TYPE IF EXISTS lounge_station_type;
-- +goose StatementEnd
