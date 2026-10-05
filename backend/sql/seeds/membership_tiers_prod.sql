BEGIN;

INSERT INTO membership_programs (program_name)
VALUES
    ('general'),
    ('ssbm')
ON CONFLICT (program_name) DO NOTHING;

WITH tier_seed(title, description, benefits, limitations, slug, program_name, group_name, stripe_product_id, is_active) AS (
    VALUES
        (
            'Day Pass',
            'Short-term access for members joining a single event or lounge visit.',
            ARRAY[
                'Full day access to Legion Lounge'
            ],
            ARRAY[]::text[],
            'day',
            'general',
            'member',
            'prod_V02wba8GuTzBju',
            TRUE
        ),
        (
            'Basic Tier',
            'The standard UBCEA membership for students and community members.',
            ARRAY[
                'Cab access',
                'Discounted raffle & ticket prices for UBCEA events',
                'Upgrade to Lounge Tier anytime & only pay the price difference'
            ],
            ARRAY[
                'No access to the Gaming Lounge'
            ],
            'basic',
            'general',
            'member',
            'prod_V02v4khwaFRJVY',
            TRUE
        ),
        (
            'Lounge Tier',
            'An upgraded membership with Lounge access and additional member perks.',
            ARRAY[
                'All Basic Tier benefits',
                'Unlimited daily 2 hour/session access to the Legion Lounge',
                'Higher discounts on raffle & ticket prices for UBCEA events'
            ],
            ARRAY[]::text[],
            'lounge',
            'general',
            'member',
            'prod_V02wxRTbZGE8m2',
            TRUE
        ),
        (
            'Competitive Player Tier',
            'Membership access for players rostered on UBCEA competitive teams.',
            ARRAY[
                'Only accessible for UBCEA competitive team players',
                'Unlimited access to the Legion Lounge'
            ],
            ARRAY[]::text[],
            'competitive_team',
            'general',
            'competitive_team',
            'prod_V02ywe6yuowM12',
            TRUE
        ),
        (
            'Executive Tier',
            'Membership access for UBCEA executives, directors, and board members.',
            ARRAY[
                'Only accessible for the UBCEA executive team'
            ],
            ARRAY[]::text[],
            'executive',
            'general',
            'executive',
            'prod_V02xxVt7nDTwbH',
            TRUE
        ),
        (
            'SSBM Semesterly',
            'A standard membership for Smash Melee events at UBCEA.',
            ARRAY[
                'Free entrance for Smash Melee weekly tournaments for a semester'
            ],
            ARRAY[
                'No discounts on Smash Melee events'
            ],
            'ssbm_semesterly',
            'ssbm',
            'member',
            'prod_VDhEsBham751Ct',
            TRUE
        ),
        (
            'SSBM Yearly',
            'A premium membership for Smash Melee events at UBCEA.',
            ARRAY[
                'Free entrance for Smash Melee weekly tournaments for the entire school year',
                'Discounts on select Smash Melee events'
            ],
            ARRAY[]::text[],
            'ssbm_yearly',
            'ssbm',
            'member',
            'prod_VDhEI2TOWQWxQ4',
            TRUE
        )
),
upserted_tiers AS (
    INSERT INTO membership_tiers (
        title,
        description,
        benefits,
        limitations,
        slug,
        program_id,
        "group",
        stripe_product_id,
        is_active,
        program_id,
        updated_at
    )
    SELECT
        title,
        description,
        benefits,
        limitations,
        slug,
        mp.id,
        group_name::group_type,
        stripe_product_id,
        is_active,
        (SELECT id FROM membership_programs WHERE program_name = 'general'),
        NOW()
    FROM tier_seed ts
    JOIN membership_programs mp
        ON mp.program_name = ts.program_name
    ON CONFLICT (stripe_product_id) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        benefits = EXCLUDED.benefits,
        limitations = EXCLUDED.limitations,
        slug = EXCLUDED.slug,
        program_id = EXCLUDED.program_id,
        "group" = EXCLUDED."group",
        is_active = EXCLUDED.is_active,
        updated_at = NOW()
    RETURNING id, slug
),
price_seed(slug, stripe_price_id, is_student_required, price_in_cents) AS (
    VALUES
        ('day', 'price_1U02qz3qdyvJ5RuGUKvK6pvI', TRUE, 500),
        ('day', 'price_1UDFrO3qdyvJ5RuGQTbalEOb', FALSE, 1000),
        ('basic', 'price_1U02pP3qdyvJ5RuGqhmcI1PM', TRUE, 1500),
        ('basic', 'price_1UDFsa3qdyvJ5RuGh7WeGx3i', FALSE, 2250),
        ('lounge', 'price_1U02qL3qdyvJ5RuGkMOYOXev', TRUE, 2500),
        ('lounge', 'price_1UDFru3qdyvJ5RuG0dnMbcie', FALSE, 3250),
        ('competitive_team', 'price_1U02s63qdyvJ5RuGoQfJ8XmQ', NULL, 0),
        ('executive', 'price_1U02rm3qdyvJ5RuGLb6JuiOO', TRUE, 1500),
        ('executive', 'price_1UDFqn3qdyvJ5RuGjOLsh4ka', FALSE, 2250),
        ('ssbm_semesterly', 'price_1UDFpy3qdyvJ5RuGKXqr0X4h', NULL, 1500),
        ('ssbm_yearly', 'price_1UDFqG3qdyvJ5RuGUr9EZpHL', NULL, 2500)
)
INSERT INTO membership_tier_prices (
    tier_id,
    stripe_price_id,
    is_student_required,
    price_in_cents,
    updated_at
)
SELECT
    t.id,
    p.stripe_price_id,
    p.is_student_required,
    p.price_in_cents,
    NOW()
FROM price_seed p
JOIN upserted_tiers t
    ON t.slug = p.slug
ON CONFLICT (stripe_price_id) DO UPDATE SET
    tier_id = EXCLUDED.tier_id,
    is_student_required = EXCLUDED.is_student_required,
    price_in_cents = EXCLUDED.price_in_cents,
    updated_at = NOW();

COMMIT;
