-- name: GetActiveTiersWithPrices :many
SELECT
    mt.id,
    mt.title,
    mt.description,
    mt.benefits,
    mt.limitations,
    mt.stripe_product_id,
    mt.slug,
    mt."group" AS required_group,
    mt.expiration_type,
    mtp.stripe_price_id,
    mtp.price_in_cents,
    mtp.is_student_required,
    mp.id AS program_id,
    mp.program_name
FROM membership_tiers mt
JOIN membership_tier_prices mtp
    ON mtp.tier_id = mt.id
JOIN membership_programs mp
    ON mp.id = mt.program_id
WHERE mt.is_active = TRUE
ORDER BY
    mp.program_name,
    mt.title;

-- name: GetPublicTiersAndPrices :many
SELECT
    mt.id,
    mt.title,
    mt.description,
    mt.benefits,
    mt.limitations,
    mt.slug,
    mt.stripe_product_id,
    mt.expiration_type,
    mtp.stripe_price_id,
    mtp.price_in_cents,
    mtp.is_student_required,
    mp.id AS program_id,
    mp.program_name
FROM membership_tiers mt
JOIN membership_tier_prices mtp
    ON mtp.tier_id = mt.id
JOIN membership_programs mp
    ON mp.id = mt.program_id
WHERE mt.is_active = TRUE AND mt."group" = 'member';

-- name: GetCurrentMembershipsWithTransactions :many
SELECT
    m.id,
    m.tier_id,
    mt.title AS tier_title,
    mt.slug,
    m.started_at,
    m.expires_at,
    m.cancelled_at,
    t.id AS transaction_id,
    t.amount_paid_cents,
    t.status,
    t.purchase_type,
    t.student_at_purchase,
    t.stripe_payment_intent_id,
    t.group_at_purchase,
    t.payment_method,
    mp.id AS program_id,
    mp.program_name
FROM memberships m
JOIN membership_tiers mt
    ON mt.id = m.tier_id
JOIN membership_programs mp
    ON mp.id = mt.program_id
JOIN transactions t
    ON t.membership_id = m.id
WHERE m.user_id = $1
    AND m.cancelled_at IS NULL
    AND m.started_at <= NOW()
    AND m.expires_at > NOW()
ORDER BY m.started_at DESC;

-- name: GetAllMembershipsWithTransactions :many
SELECT
    m.id,
    m.tier_id,
    mt.title AS tier_title,
    mt.slug,
    m.started_at,
    m.expires_at,
    m.cancelled_at,
    t.id AS transaction_id,
    t.amount_paid_cents,
    t.status,
    t.student_at_purchase,
    t.purchase_type,
    t.stripe_payment_intent_id,
    t.group_at_purchase,
    t.payment_method,
    mp.id AS program_id,
    mp.program_name
FROM memberships m
JOIN membership_tiers mt
    ON mt.id = m.tier_id
JOIN membership_programs mp
    ON mp.id = mt.program_id
JOIN transactions t
    ON t.membership_id = m.id
WHERE m.user_id = $1
ORDER BY m.started_at DESC;


-- name: GetTierByTierId :one
SELECT
    mt.id,
    mt.title,
    mt.description,
    mt.benefits,
    mt.limitations,
    mt.slug,
    mt.stripe_product_id,
    mt.expiration_type,
    mtp.stripe_price_id,
    mtp.price_in_cents,
    mtp.is_student_required,
    mp.id AS program_id,
    mp.program_name
FROM membership_tiers mt
JOIN membership_tier_prices mtp
    ON mtp.tier_id = mt.id
JOIN membership_programs mp
    ON mp.id = mt.program_id
WHERE mt.id = $1;


-- name: CreateMembership :one
INSERT INTO memberships (
    user_id,
    tier_id,
    started_at,
    expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id;

-- name: CancelActiveMembershipByUserIdAndMembershipId :execrows
UPDATE memberships
SET
    cancelled_at = $3,
    updated_at = $3
WHERE user_id = $1
    AND id = $2
    AND cancelled_at IS NULL
    AND started_at <= NOW()
    AND expires_at > NOW();

-- name: CancelActiveMembershipsByUserIdAndProgramId :exec
UPDATE memberships AS m
SET
    cancelled_at = $3,
    updated_at = $3
FROM membership_tiers AS mt
WHERE mt.id = m.tier_id
    AND m.user_id = $1
    AND mt.program_id = $2
    AND m.cancelled_at IS NULL
    AND m.started_at <= NOW()
    AND m.expires_at > NOW();




-- Transaction related stuff

-- name: GetPendingTransactionForUpdate :one
SELECT
    id,
    stripe_checkout_session_id
FROM transactions
WHERE user_id = $1 AND status = 'pending'
FOR UPDATE;

-- name: ExpirePendingTransactionById :exec
UPDATE transactions
SET
    status = 'expired',
    updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- name: CreatePendingTransaction :one
INSERT INTO transactions (
    user_id,
    tier_id,
    group_at_purchase,
    student_at_purchase,
    purchase_type,
    payment_method,
    status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    'pending'
)
RETURNING id;

-- name: PutStripeCheckoutSessionId :execrows
UPDATE transactions
SET
    stripe_checkout_session_id = $2,
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending';

-- name: UpdateTransactionStatusById :exec
UPDATE transactions
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdatePendingTransactionStatusByCheckoutId :exec
UPDATE transactions
SET
    status = $2,
    updated_at = NOW()
WHERE stripe_checkout_session_id = $1 AND status = 'pending';

-- name: GetTransactionByCheckoutSessionIdForUpdate :one
SELECT
    t.id,
    t.user_id,
    t.membership_id,
    t.tier_id,
    t.status,
    t.purchase_type,
    t.stripe_checkout_session_id,
    mt.program_id,
    mt.expiration_type
FROM transactions AS t
JOIN membership_tiers AS mt
    ON mt.id = t.tier_id
WHERE t.stripe_checkout_session_id = $1
FOR UPDATE OF t;

-- name: GetTransactionByTransactionIdForUpdate :one
SELECT
    t.id,
    t.user_id,
    t.membership_id,
    t.tier_id,
    t.status,
    t.purchase_type,
    t.stripe_checkout_session_id,
    mt.program_id,
    mt.expiration_type
FROM transactions AS t
JOIN membership_tiers AS mt
    ON mt.id = t.tier_id
WHERE t.id = $1
FOR UPDATE OF t;

-- name: CompleteTransaction :exec
UPDATE transactions
SET
    membership_id = $2,
    stripe_payment_intent_id = $3,
    amount_paid_cents = $4,
    status = 'completed',
    updated_at = NOW()
WHERE id = $1 AND status = 'pending';

-- name: GetActiveMembershipsExpiringOnDate :many
-- Active (non-cancelled) memberships whose expires_at falls on the given
-- calendar date in Vancouver time. Used by the daily expiry-notification job:
-- called with (expiry date - 7 days) for "expiring soon" emails, and with
-- today's date for "expired" emails.
SELECT
    m.id,
    m.user_id,
    m.expires_at,
    mt.title AS tier_title,
    u.email
FROM memberships m
JOIN membership_tiers mt ON mt.id = m.tier_id
JOIN users u ON u.id = m.user_id
WHERE m.cancelled_at IS NULL
    AND (m.expires_at AT TIME ZONE 'America/Vancouver')::date = $1::date;
