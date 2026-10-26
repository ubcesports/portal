BEGIN;

-- Bulk sample data for exercising the admin analytics page (/admin/analytics)
-- with more than a single data point: 16 mock users purchasing across
-- tiers, student status, and purchase types, spread over the last ~8 months
-- so the weekly/monthly/yearly charts each show several buckets, and
-- "since inception" shows meaningfully more history than "last N periods".

DELETE FROM transactions
WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'mock.analytics.%@example.com');

DELETE FROM memberships
WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'mock.analytics.%@example.com');

DELETE FROM users WHERE email LIKE 'mock.analytics.%@example.com';

WITH sample(n, full_name, student_id, is_student, tier_slug, days_ago, amount_cents, purchase_type) AS (
    VALUES
        (1,  'Ava Chen',      'MOCKA001', TRUE,  'day',              2,   500,  'new'),
        (2,  'Ben Okafor',    NULL,       FALSE, 'day',              5,   1000, 'new'),
        (3,  'Carla Reyes',   'MOCKA003', TRUE,  'basic',            9,   1500, 'new'),
        (4,  'Dev Patel',     NULL,       FALSE, 'basic',            14,  2250, 'new'),
        (5,  'Ella Novak',    'MOCKA005', TRUE,  'lounge',           14,  2500, 'upgrade'),
        (6,  'Finn Walsh',    NULL,       FALSE, 'lounge',           20,  3250, 'new'),
        (7,  'Grace Kim',     'MOCKA007', TRUE,  'basic',            25,  1500, 'new'),
        (8,  'Hugo Alvarez',  NULL,       FALSE, 'executive',        30,  2250, 'new'),
        (9,  'Ivy Thompson',  'MOCKA009', TRUE,  'competitive_team', 33,  0,    'new'),
        (10, 'Jae Park',      NULL,       FALSE, 'day',              40,  1000, 'new'),
        (11, 'Kira Singh',    'MOCKA011', TRUE,  'basic',            48,  1500, 'new'),
        (12, 'Leo Martins',   NULL,       FALSE, 'lounge',           55,  3250, 'upgrade'),
        (13, 'Mira Haddad',   'MOCKA013', TRUE,  'basic',            63,  1500, 'new'),
        (14, 'Noah Fischer',  NULL,       FALSE, 'day',              70,  1000, 'new'),
        (15, 'Omi Tanaka',    'MOCKA015', TRUE,  'lounge',           160, 2500, 'new'),
        (16, 'Priya Nair',    NULL,       FALSE, 'basic',            250, 2250, 'new')
),
inserted_users AS (
    INSERT INTO users (
        email,
        student_id,
        role,
        created_at,
        updated_at,
        full_name,
        email_verified_at,
        is_student,
        onboarding_completed_at
    )
    SELECT
        'mock.analytics.' || n || '@example.com',
        student_id,
        'member',
        NOW() - make_interval(days => days_ago + 1),
        NOW(),
        full_name,
        NOW() - make_interval(days => days_ago + 1),
        is_student,
        NOW() - make_interval(days => days_ago)
    FROM sample
    RETURNING id, email
),
inserted_memberships AS (
    INSERT INTO memberships (user_id, tier_id, started_at, expires_at)
    SELECT
        u.id,
        mt.id,
        NOW() - make_interval(days => s.days_ago),
        NOW() - make_interval(days => s.days_ago) + INTERVAL '365 days'
    FROM sample s
    JOIN inserted_users u ON u.email = 'mock.analytics.' || s.n || '@example.com'
    JOIN membership_tiers mt ON mt.slug = s.tier_slug
    RETURNING id, user_id
)
INSERT INTO transactions (
    user_id,
    membership_id,
    tier_id,
    stripe_payment_intent_id,
    status,
    group_at_purchase,
    student_at_purchase,
    amount_paid_cents,
    purchase_type,
    created_at
)
SELECT
    im.user_id,
    im.id,
    mt.id,
    'seed_mock_pi_analytics_' || s.n,
    'completed'::transaction_status_type,
    (CASE s.tier_slug
        WHEN 'competitive_team' THEN 'competitive_team'
        WHEN 'executive' THEN 'executive'
        ELSE 'member'
    END)::group_type,
    s.is_student,
    s.amount_cents,
    s.purchase_type::purchase_type,
    NOW() - make_interval(days => s.days_ago)
FROM sample s
JOIN inserted_users u ON u.email = 'mock.analytics.' || s.n || '@example.com'
JOIN inserted_memberships im ON im.user_id = u.id
JOIN membership_tiers mt ON mt.slug = s.tier_slug;

COMMIT;
