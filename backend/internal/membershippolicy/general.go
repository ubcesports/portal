package membershippolicy

import (
	"slices"

	"github.com/ubcesports/memberships/internal/dto"
)

type GeneralPolicy struct {
	standard *TransitionPolicy
}

func NewGeneralPolicy() Policy {
	return &GeneralPolicy{
		standard: NewTransitionPolicy(
			"general",
			map[Transition]dto.PurchaseType{
				{From: "", To: "day"}:    dto.PurchaseNew,
				{From: "", To: "basic"}:  dto.PurchaseNew,
				{From: "", To: "lounge"}: dto.PurchaseNew,

				{From: "day", To: "basic"}:  dto.PurchaseNew,
				{From: "day", To: "lounge"}: dto.PurchaseNew,

				{
					From: "basic",
					To:   "lounge",
				}: dto.PurchaseUpgrade,
			},
		),
	}
}

func (p *GeneralPolicy) ProgramName() string {
	return "general"
}

func (p *GeneralPolicy) Evaluate(
	profile *dto.ProfileDTO,
	current *dto.MembershipDTO,
	requested *dto.MembershipTierDTO,
) (EvaluationResult, error) {
	// Exec groups have the highest priority inside general
	if slices.Contains(profile.Groups, dto.GroupExecutive) ||
		slices.Contains(profile.Groups, dto.GroupDirector) ||
		slices.Contains(profile.Groups, dto.GroupBoard) {
		return allowRestrictedGeneralTier(
			current,
			requested,
			"executive",
			dto.ReasonExecutiveRestricted,
		), nil
	}

	// Comp players is the next priority
	if slices.Contains(profile.Groups, dto.GroupCompetitiveTeam) {
		return allowRestrictedGeneralTier(
			current,
			requested,
			"competitive_team",
			dto.ReasonCompetitiveRestricted,
		), nil
	}

	// Regular members follow day/basic/lounge transitions
	return p.standard.Evaluate(
		profile,
		current,
		requested,
	)
}

// allowRestrictedGeneralTier gates a tier restricted to a specific group
// (eg. executive, competitive_team): the group's members may only ever hold
// their own dedicated tier, one at a time. Any other tier in this program is
// unavailable to them, tagged with the given reason regardless of why.
func allowRestrictedGeneralTier(
	current *dto.MembershipDTO,
	requested *dto.MembershipTierDTO,
	requiredSlug string,
	restrictedReason dto.TierUnavailableReason,
) EvaluationResult {
	if requested.Slug != requiredSlug {
		return EvaluationResult{Reason: restrictedReason}
	}

	if current != nil {
		if current.Slug == requested.Slug {
			return EvaluationResult{Reason: dto.ReasonAlreadyOwned}
		}
		// Holds a different tier in this program (eg. a "day" pass from
		// before joining the group) — can't also hold this one.
		return EvaluationResult{Reason: dto.ReasonNotEligibleCurrent}
	}

	return EvaluationResult{PurchaseType: dto.PurchaseNew, Allowed: true}
}
