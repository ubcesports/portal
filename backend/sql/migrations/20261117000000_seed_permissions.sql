-- +goose Up
-- +goose StatementBegin
-- Keep in sync with internal/auth/permissions.go (TestPermissionsMatchMigrationAndSeed).
INSERT INTO permissions (permission)
VALUES
    ('users.view'),
    ('users.update'),
    ('users.export'),
    ('membership_tiers.view'),
    ('memberships.view'),
    ('memberships.create'),
    ('memberships.delete'),
    ('membership_invitations.view'),
    ('membership_invitations.create'),
    ('membership_invitations.delete'),
    ('exec_profiles.update'),
    ('audit_logs.view'),
    ('audit_logs.export'),
    ('analytics.view'),
    ('analytics.export'),
    ('lounge_sessions.create'),
    ('lounge_sessions.checkout'),
    ('lounge_sessions.view'),
    ('lounge_stations.update')
ON CONFLICT (permission) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions;
DELETE FROM permissions;
-- +goose StatementEnd
