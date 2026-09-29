package dto

import "time"

// Membership tiers

type MembershipTierDTO struct {
	ID             string                   `json:"id"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	Slug           string                   `json:"slug"`
	ProductId      string                   `json:"product_id"`
	Benefits       []string                 `json:"benefits"`
	Limitations    []string                 `json:"limitations"`
	Prices         []MembershipTierPriceDTO `json:"prices"`
	ProgramId      string                   `json:"program_id"`
	ProgramName    string                   `json:"program_name"`
	ExpirationType MembershipExpirationType `json:"expiration_type"`
	RequiredGroup  GroupType                `json:"-"`
}

type EligibleMembershipTierDTO struct {
	ID             string                   `json:"id"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	Slug           string                   `json:"slug"`
	ProductId      string                   `json:"product_id"`
	Benefits       []string                 `json:"benefits"`
	Limitations    []string                 `json:"limitations"`
	ProgramId      string                   `json:"program_id"`
	ProgramName    string                   `json:"program_name"`
	ExpirationType MembershipExpirationType `json:"expiration_type"`

	Eligible          bool                    `json:"eligible"`
	PurchaseType      PurchaseType            `json:"purchase_type,omitempty"`
	Price             *MembershipTierPriceDTO `json:"prices,omitempty"`
	UnavailableReason TierUnavailableReason   `json:"unavailable_reason,omitempty"`
	PurchaseOpensAt   *time.Time              `json:"purchase_opens_at,omitempty"`
}

type MembershipTierPriceDTO struct {
	Price             float64 `json:"price"`
	PriceId           string  `json:"price_id"`
	IsStudentRequired *bool   `json:"is_student_required"`
}

// Memberships

type MembershipDTO struct {
	ID          string         `json:"id"`
	TierId      string         `json:"tier_id"`
	TierTitle   string         `json:"tier_title"`
	Slug        string         `json:"slug"`
	StartedAt   time.Time      `json:"started_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	CancelledAt *time.Time     `json:"cancelled_at"`
	Transaction TransactionDTO `json:"transaction"`
	ProgramId   string         `json:"program_id"`
	ProgramName string         `json:"program_name"`
}

type TransactionDTO struct {
	ID                    string                `json:"id"`
	AmountPaid            string                `json:"amount_paid"`
	Status                TransactionStatusType `json:"status"`
	GroupAtPurchase       GroupType             `json:"group_at_purchase"`
	PurchaseType          PurchaseType          `json:"purchase_type"`
	StripePaymentIntentId string                `json:"stripe_payment_intent_id"`
	StudentAtPurchase     bool                  `json:"student_at_purchase"`
	PaymentMethod         PaymentMethodType     `json:"payment_method"`
	AmountPaidCents       int64                 `json:"-"`
}

// Request

type CheckoutSessionRequest struct {
	TierId string `json:"tier_id"`
}

// Response

type CheckoutSessionResponse struct {
	Url string `json:"url"`
}
