BEGIN;

-- Mock lounge members and staff for local testing. Run after migrations and the
-- tier and lounge station seeds. Safe to re-run.
--
-- Lounge members (student ID: case)
--   MOCK2000 Leo Lounge       active `lounge` tier        -> can check in
--   MOCK2001 Alex Day         active `day` tier           -> can check in (second member)
--   MOCK2002 Sam Basic        active `basic` tier only    -> no lounge access
--   MOCK2003 Casey Cancelled  `lounge` tier, cancelled    -> no lounge access
--   MOCK2004 Pat Future       `lounge` tier, starts in 5d -> no lounge access yet
--   MOCK2005 Morgan Multi     `basic` + `executive`       -> eligible through executive
--
-- Staff (no student ID), each holding one custom role
--   lounge.staff.mock@example.com  Lena Lounge  -> "Lounge Staff"
--   finance.mock@example.com       Finn Finance -> "Treasurer"
--   norole.mock@example.com        Nora NoRole  -> no custom role

INSERT INTO users (
    email,
    student_id,
    role,
    full_name,
    is_student,
    email_verified_at,
    onboarding_completed_at
)
VALUES
    ('leo.lounge.mock@example.com',        'MOCK2000', 'member', 'Leo Lounge',      TRUE,  NOW(), NOW()),
    ('alex.day.mock@example.com',          'MOCK2001', 'member', 'Alex Day',        TRUE,  NOW(), NOW()),
    ('sam.basic.mock@example.com',         'MOCK2002', 'member', 'Sam Basic',       TRUE,  NOW(), NOW()),
    ('casey.cancelled.mock@example.com',   'MOCK2003', 'member', 'Casey Cancelled', TRUE,  NOW(), NOW()),
    ('pat.future.mock@example.com',        'MOCK2004', 'member', 'Pat Future',      TRUE,  NOW(), NOW()),
    ('morgan.multi.mock@example.com',      'MOCK2005', 'member', 'Morgan Multi',    TRUE,  NOW(), NOW()),
    ('lounge.staff.mock@example.com',      NULL,       'member', 'Lena Lounge',     FALSE, NOW(), NOW()),
    ('finance.mock@example.com',           NULL,       'member', 'Finn Finance',    FALSE, NOW(), NOW()),
    ('norole.mock@example.com',            NULL,       'member', 'Nora NoRole',     FALSE, NOW(), NOW())
ON CONFLICT (email) DO UPDATE SET
    student_id = EXCLUDED.student_id,
    role = EXCLUDED.role,
    full_name = EXCLUDED.full_name,
    is_student = EXCLUDED.is_student,
    email_verified_at = EXCLUDED.email_verified_at,
    onboarding_completed_at = EXCLUDED.onboarding_completed_at,
    updated_at = NOW();

-- Memberships. Cleared first so the seed stays rerunnable.
DELETE FROM lounge_sessions
WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%.mock@example.com');

DELETE FROM memberships
WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%.mock@example.com');

INSERT INTO memberships (user_id, tier_id, started_at, expires_at, cancelled_at)
SELECT u.id, mt.id, NOW() + m.starts, NOW() + m.expires, m.cancelled
FROM (
    VALUES
        ('leo.lounge.mock@example.com',      'lounge',    INTERVAL '-1 day', INTERVAL '30 days', NULL::timestamptz),
        ('alex.day.mock@example.com',        'day',       INTERVAL '-1 day', INTERVAL '30 days', NULL),
        ('sam.basic.mock@example.com',       'basic',     INTERVAL '-1 day', INTERVAL '30 days', NULL),
        ('casey.cancelled.mock@example.com', 'lounge',    INTERVAL '-1 day', INTERVAL '30 days', NOW()),
        ('pat.future.mock@example.com',      'lounge',    INTERVAL '5 days', INTERVAL '35 days', NULL),
        ('morgan.multi.mock@example.com',    'basic',     INTERVAL '-1 day', INTERVAL '30 days', NULL),
        ('morgan.multi.mock@example.com',    'executive', INTERVAL '-1 day', INTERVAL '30 days', NULL)
) AS m (email, tier_slug, starts, expires, cancelled)
JOIN users u ON u.email = m.email
JOIN membership_tiers mt ON mt.slug = m.tier_slug;

-- Custom roles and the permissions each grants.
INSERT INTO roles (name, description, color, display_order)
VALUES
    ('Lounge Staff', 'Runs the lounge front desk: check-in, check-out and station control.', '#2563EB', 10),
    ('Treasurer',    'Views members and memberships, and views and exports analytics.',      '#16A34A', 20)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    color = EXCLUDED.color,
    display_order = EXCLUDED.display_order,
    updated_at = NOW();

DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('Lounge Staff', 'Treasurer'));

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, rp.permission
FROM (
    VALUES
        ('Lounge Staff', 'lounge_sessions.create'),
        ('Lounge Staff', 'lounge_sessions.checkout'),
        ('Lounge Staff', 'lounge_sessions.view'),
        ('Lounge Staff', 'lounge_stations.update'),
        ('Treasurer',    'users.view'),
        ('Treasurer',    'memberships.view'),
        ('Treasurer',    'analytics.view'),
        ('Treasurer',    'analytics.export')
) AS rp (role_name, permission)
JOIN roles r ON r.name = rp.role_name;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM (
    VALUES
        ('lounge.staff.mock@example.com', 'Lounge Staff'),
        ('finance.mock@example.com',      'Treasurer')
) AS ur (email, role_name)
JOIN users u ON u.email = ur.email
JOIN roles r ON r.name = ur.role_name
ON CONFLICT (user_id, role_id) DO NOTHING;

COMMIT;
