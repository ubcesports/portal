package dto

import "time"

type MembershipInvitationDTO struct {
	ID               string            `json:"id"`
	Email            string            `json:"email"`
	TierID           string            `json:"tier_id"`
	TierTitle        string            `json:"tier_title"`
	ProgramName      string            `json:"program_name"`
	AmountPaidCents  int64             `json:"amount_paid_cents"`
	PaymentMethod    PaymentMethodType `json:"payment_method"`
	Done             bool              `json:"done"`
	CreatedByUserID  string            `json:"created_by_user_id"`
	CreatedByName    string            `json:"created_by_name"`
	InvitationSentAt *time.Time        `json:"invitation_sent_at"`
	PurchasedAt      time.Time         `json:"purchased_at"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type CreateMembershipInvitationRequest struct {
	Email           string            `json:"email"`
	TierID          string            `json:"tier_id"`
	AmountPaidCents int64             `json:"amount_paid_cents"`
	PaymentMethod   PaymentMethodType `json:"payment_method"`
}
