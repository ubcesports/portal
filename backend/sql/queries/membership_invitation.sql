-- name: CreateMembershipInvitation :one
WITH inserted AS (
    INSERT INTO membership_invitations (
        email,
        tier_id,
        amount_paid_cents,
        payment_method,
        created_by_user_id,
        invitation_sent_at
    ) VALUES (
        sqlc.arg(email),
        sqlc.arg(tier_id),
        sqlc.arg(amount_paid_cents),
        sqlc.arg(payment_method),
        sqlc.arg(created_by_user_id),
        NOW()
    )
    RETURNING *
)
SELECT
    p.id,
    p.email,
    p.tier_id,
    mt.title AS tier_title,
    mp.program_name,
    p.amount_paid_cents,
    p.payment_method,
    p.done,
    p.created_by_user_id,
    creator.full_name AS created_by_name,
    p.invitation_sent_at,
    p.purchased_at,
    p.created_at,
    p.updated_at
FROM inserted p
JOIN membership_tiers mt ON mt.id = p.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
JOIN users creator ON creator.id = p.created_by_user_id;

-- name: GetMembershipInvitations :many
SELECT
    p.id,
    p.email,
    p.tier_id,
    mt.title AS tier_title,
    mp.program_name,
    p.amount_paid_cents,
    p.payment_method,
    p.done,
    p.created_by_user_id,
    creator.full_name AS created_by_name,
    p.invitation_sent_at,
    p.purchased_at,
    p.created_at,
    p.updated_at
FROM membership_invitations p
JOIN membership_tiers mt ON mt.id = p.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
JOIN users creator ON creator.id = p.created_by_user_id
ORDER BY p.created_at DESC, p.id DESC;

-- name: GetMembershipInvitationByID :one
SELECT
    p.id,
    p.email,
    p.tier_id,
    mt.title AS tier_title,
    mp.program_name,
    p.amount_paid_cents,
    p.payment_method,
    p.done,
    p.created_by_user_id,
    creator.full_name AS created_by_name,
    p.invitation_sent_at,
    p.purchased_at,
    p.created_at,
    p.updated_at
FROM membership_invitations p
JOIN membership_tiers mt ON mt.id = p.tier_id
JOIN membership_programs mp ON mp.id = mt.program_id
JOIN users creator ON creator.id = p.created_by_user_id
WHERE p.id = sqlc.arg(id);

-- name: DeleteMembershipInvitation :execrows
DELETE FROM membership_invitations
WHERE id = sqlc.arg(id);

-- name: MembershipInvitationUserExistsByEmail :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE LOWER(email) = LOWER(sqlc.arg(email))
);

-- name: GetMembershipInvitationByEmailForUpdate :one
SELECT
    p.id,
    p.tier_id,
    p.amount_paid_cents,
    p.payment_method,
    p.purchased_at,
    mt.program_id,
    mt.expiration_type
FROM membership_invitations p
JOIN membership_tiers mt ON mt.id = p.tier_id
WHERE LOWER(p.email) = LOWER(sqlc.arg(email))
  AND p.done = FALSE
FOR UPDATE OF p;

-- name: MarkMembershipInvitationDone :execrows
UPDATE membership_invitations
SET
    done = TRUE,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND done = FALSE;
