-- Permanently delete a user and every row associated with them.
--
-- Replace the UUID below before running this file. Everything runs in one
-- transaction, so any error rolls the entire deletion back.

BEGIN;

DO $delete_user$
DECLARE
    user_id_to_delete CONSTANT UUID := '00000000-0000-0000-0000-000000000000';
    target_email VARCHAR;
BEGIN
    SELECT email
    INTO STRICT target_email
    FROM users
    WHERE id = user_id_to_delete;

    -- Audit rows may reference the user as either the actor or the target.
    DELETE FROM admin_audit_logs
    WHERE actor_user_id = user_id_to_delete
       OR target_user_id = user_id_to_delete;

    DELETE FROM membership_invitations
    WHERE created_by_user_id = user_id_to_delete;

    -- exec_social_link rows are removed by exec_profile's ON DELETE CASCADE.
    DELETE FROM exec_profile
    WHERE user_id = user_id_to_delete;

    DELETE FROM transactions
    WHERE user_id = user_id_to_delete
       OR membership_id IN (
           SELECT id
           FROM memberships
           WHERE user_id = user_id_to_delete
       );

    DELETE FROM memberships
    WHERE user_id = user_id_to_delete;

    DELETE FROM user_groups
    WHERE user_id = user_id_to_delete;

    DELETE FROM sessions
    WHERE user_id = user_id_to_delete;

    DELETE FROM accounts
    WHERE user_id = user_id_to_delete;

    -- Limen stores subjects as "action::identifier", with the user's email
    -- as the identifier for user-specific verification records.
    DELETE FROM verifications
    WHERE split_part(subject, '::', 2) = target_email;

    DELETE FROM users
    WHERE id = user_id_to_delete;

    RAISE NOTICE 'Deleted user % (%) and all associated data',
        user_id_to_delete,
        target_email;
EXCEPTION
    WHEN NO_DATA_FOUND THEN
        RAISE EXCEPTION 'No user exists with id %', user_id_to_delete;
    WHEN TOO_MANY_ROWS THEN
        RAISE EXCEPTION 'More than one user exists with id %', user_id_to_delete;
END
$delete_user$;

COMMIT;
