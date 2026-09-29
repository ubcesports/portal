package membershippolicy

import (
	"testing"

	"github.com/ubcesports/memberships/internal/dto"
)

func TestTransitionPolicyRejectsWrongProgram(t *testing.T) {
	policy := NewSSBMPolicy()
	tier := &dto.MembershipTierDTO{Slug: "ssbm_yearly", ProgramName: "general"}

	_, err := policy.Evaluate(&dto.ProfileDTO{}, nil, tier)
	if err == nil {
		t.Fatal("expected an error when evaluating a tier from a different program")
	}
}

func TestSSBMPolicyTransitions(t *testing.T) {
	policy := NewSSBMPolicy()
	tier := func(slug string) *dto.MembershipTierDTO {
		return &dto.MembershipTierDTO{Slug: slug, ProgramName: "ssbm"}
	}
	current := func(slug string) *dto.MembershipDTO {
		return &dto.MembershipDTO{Slug: slug}
	}

	tests := []struct {
		name        string
		current     *dto.MembershipDTO
		requested   string
		wantAllowed bool
		wantReason  dto.TierUnavailableReason
	}{
		{"no membership -> semesterly", nil, "ssbm_semesterly", true, ""},
		{"no membership -> yearly", nil, "ssbm_yearly", true, ""},
		{"owns semesterly -> yearly is an upgrade", current("ssbm_semesterly"), "ssbm_yearly", true, ""},
		{"owns semesterly -> semesterly is already owned", current("ssbm_semesterly"), "ssbm_semesterly", false, dto.ReasonAlreadyOwned},
		{"owns yearly -> yearly is already owned", current("ssbm_yearly"), "ssbm_yearly", false, dto.ReasonAlreadyOwned},
		{"owns yearly -> semesterly is not eligible", current("ssbm_yearly"), "ssbm_semesterly", false, dto.ReasonNotEligibleCurrent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := policy.Evaluate(&dto.ProfileDTO{}, tt.current, tier(tt.requested))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Allowed != tt.wantAllowed {
				t.Fatalf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Fatalf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}
