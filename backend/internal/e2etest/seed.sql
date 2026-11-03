-- Used only by scripts/e2e.sh against its disposable Compose database.
\set ON_ERROR_STOP on
DO $$ BEGIN
  IF current_database() <> 'portal_e2e' OR current_user <> 'portal_e2e' THEN
    RAISE EXCEPTION 'Refusing to reset a non-E2E database';
  END IF;
END $$;
BEGIN;
TRUNCATE users, membership_programs, verifications CASCADE;

INSERT INTO membership_programs (id, program_name)
VALUES ('10000000-0000-0000-0000-000000000001', 'general');
INSERT INTO membership_tiers
(id, title, description, benefits, limitations, slug, program_id, "group", expiration_type, stripe_product_id, is_active)
VALUES
('20000000-0000-0000-0000-000000000001', 'Basic Tier', 'E2E basic membership', ARRAY['E2E member benefits'], ARRAY['No lounge access'], 'basic', '10000000-0000-0000-0000-000000000001', 'member', 'year', 'prod_e2e_basic', true),
('20000000-0000-0000-0000-000000000002', 'Lounge Tier', 'E2E lounge membership', ARRAY['E2E lounge access'], ARRAY[]::text[], 'lounge', '10000000-0000-0000-0000-000000000001', 'member', 'year', 'prod_e2e_lounge', true),
('20000000-0000-0000-0000-000000000003', 'Day Pass', 'E2E day membership', ARRAY['E2E daily access'], ARRAY[]::text[], 'day', '10000000-0000-0000-0000-000000000001', 'member', 'day', 'prod_e2e_day', true);
INSERT INTO membership_tier_prices (tier_id, stripe_price_id, is_student_required, price_in_cents)
SELECT id, 'price_e2e_' || slug || '_student', true, 500 FROM membership_tiers;
INSERT INTO membership_tier_prices (tier_id, stripe_price_id, is_student_required, price_in_cents)
SELECT id, 'price_e2e_' || slug || '_community', false, 1000 FROM membership_tiers;

INSERT INTO users (id, email, full_name, role, is_student, onboarding_completed_at, email_verified_at)
VALUES
('30000000-0000-0000-0000-000000000001', 'member@e2e.invalid', 'E2E Member', 'member', false, now(), now()),
('30000000-0000-0000-0000-000000000002', 'admin@e2e.invalid', 'E2E Administrator', 'admin', false, now(), now()),
('30000000-0000-0000-0000-000000000003', 'new@e2e.invalid', 'E2E New', 'member', false, NULL, now()),
('30000000-0000-0000-0000-000000000004', 'history@e2e.invalid', 'E2E History', 'member', false, now(), now());
INSERT INTO accounts (user_id, provider, provider_account_id, access_token, scope, created_at, updated_at)
SELECT id, 'zetrova', split_part(email, '@', 1), 'e2e-placeholder', 'openid profile email', now(), now() FROM users;
INSERT INTO user_groups (user_id, "group") SELECT id, 'member' FROM users WHERE onboarding_completed_at IS NOT NULL;

INSERT INTO memberships (id, user_id, tier_id, started_at, expires_at)
VALUES
('40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000001', now() - interval '1 day', now() + interval '30 days'),
('40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000003', now() - interval '60 days', now() - interval '59 days');
INSERT INTO transactions (user_id, membership_id, tier_id, status, group_at_purchase, student_at_purchase, amount_paid_cents, purchase_type, payment_method)
SELECT user_id, id, tier_id, 'completed', 'member', false, 1000, 'new', 'cash' FROM memberships;
COMMIT;
