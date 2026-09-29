package membershippolicy

import "github.com/ubcesports/memberships/internal/dto"

// EvaluationResult is a policy's verdict on whether a specific tier is
// currently purchasable for a user. Reason is only meaningful when Allowed
// is false, and explains why for display on the pricing page.
type EvaluationResult struct {
	PurchaseType dto.PurchaseType
	Allowed      bool
	Reason       dto.TierUnavailableReason
}

type Policy interface {
	ProgramName() string

	Evaluate(
		profile *dto.ProfileDTO,
		current *dto.MembershipDTO,
		requested *dto.MembershipTierDTO,
	) (EvaluationResult, error)
}
