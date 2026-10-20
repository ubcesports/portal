-- +goose Up
-- +goose StatementBegin
CREATE TABLE membership_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR NOT NULL,
    tier_id UUID NOT NULL REFERENCES membership_tiers(id),
    amount_paid_cents BIGINT NOT NULL,
    payment_method payment_method_type NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id UUID NOT NULL REFERENCES users(id),
    invitation_sent_at TIMESTAMPTZ,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT membership_invitations_email_not_blank
        CHECK (BTRIM(email) <> ''),
    CONSTRAINT membership_invitations_amount_paid_nonnegative
        CHECK (amount_paid_cents >= 0),
    CONSTRAINT membership_invitations_offline_payment_method
        CHECK (payment_method IN ('cash', 'etransfer'))
);

CREATE UNIQUE INDEX membership_invitations_email_unique
    ON membership_invitations (LOWER(email));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS membership_invitations;
-- +goose StatementEnd
