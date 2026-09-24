BEGIN;

WITH tier_seed(title, description, benefits, slug, group_name, stripe_product_id, is_active) AS (
    VALUES
        (
            'Day Pass',
            'Short-term access for members joining a single event or lounge visit.',
            ARRAY[
                'Full day access to Legion Lounge'
            ],
            'day',
            'member',
            'prod_UlEQ2pcezb8VkF',
            TRUE
        ),
        (
            'Basic Tier',
            'The standard UBCEA membership for students and community members.',
            ARRAY[
                'No access to the Legion Lounge',
                'Cab access',
                'Discounted raffle & ticket prices for UBCEA events',
                'Upgrade to Lounge Tier anytime & only pay the price difference'
            ],
            'basic',
            'member',
            'prod_UnGYL5Cxq8Sr7r',
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
            'lounge',
            'member',
            'prod_UnGYUlfwqi6af8',
            TRUE
        ),
        (
            'Competitive Player Tier',
            'Membership access for players rostered on UBCEA competitive teams.',
            ARRAY[
                'Only accessible for UBCEA competitive team players',
                'Unlimited access to the Legion Lounge'
            ],
            'competitive_team',
            'competitive_team',
            'prod_UnGZnjFGP74y44',
            TRUE
        ),
        (
            'Executive Tier',
            'Membership access for UBCEA executives, directors, and board members.',
            ARRAY[
                'Only accessible for the UBCEA executive team'
            ],
            'executive',
            'executive',
            'prod_UnGZIuiCafLoqL',
            TRUE
        )
),
upserted_tiers AS (
    INSERT INTO membership_tiers (
        title,
        description,
        benefits,
        slug,
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
        slug,
        group_name::group_type,
        stripe_product_id,
        is_active,
        (SELECT id FROM membership_programs WHERE program_name = 'general'),
        NOW()
    FROM tier_seed
    ON CONFLICT (stripe_product_id) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        benefits = EXCLUDED.benefits,
        slug = EXCLUDED.slug,
        "group" = EXCLUDED."group",
        is_active = EXCLUDED.is_active,
        updated_at = NOW()
    RETURNING id, slug
),
price_seed(slug, stripe_price_id, is_student_required, price_in_cents) AS (
    VALUES
        ('day', 'price_1TlhxGDEhs9s474KiqWxmkwk', TRUE, 500),
        ('day', 'price_1TlhxaDEhs9s474KjGKdv4Zl', FALSE, 1000),
        ('basic', 'price_1Tng1SDEhs9s474KmB0chFf4', TRUE, 1500),
        ('basic', 'price_1Tng2XDEhs9s474KeFwj2xfZ', FALSE, 2250),
        ('lounge', 'price_1Tng1hDEhs9s474KJUoSaOy8', TRUE, 2500),
        ('lounge', 'price_1Tng30DEhs9s474KWdiXE6iK', FALSE, 3250),
        ('competitive_team', 'price_1Tng27DEhs9s474KStx6hYtV', NULL, 0),
        ('executive', 'price_1Tng1sDEhs9s474KxHto1E80', TRUE, 1500),
        ('executive', 'price_1Tng3TDEhs9s474K0MpRQlaG', FALSE, 2250)
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
