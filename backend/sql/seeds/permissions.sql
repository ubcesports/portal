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
