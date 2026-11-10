-- +goose Up
-- +goose StatementBegin
-- Permission keys must match the constants in internal/auth/permissions.go;
-- TestPermissionsMatchMigration enforces that. Add new keys in a new migration.
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
    ('analytics.export')
ON CONFLICT (permission) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM role_permissions;
DELETE FROM permissions;
-- +goose StatementEnd
