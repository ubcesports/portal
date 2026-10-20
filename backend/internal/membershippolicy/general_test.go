package membershippolicy

import (
	"testing"

	"github.com/ubcesports/memberships/internal/dto"
)

func generalTier(slug string) *dto.MembershipTierDTO {
	return &dto.MembershipTierDTO{ID: slug + "-id", Slug: slug, ProgramName: "general"}
}

func membershipOf(slug string) *dto.MembershipDTO {
	return &dto.MembershipDTO{Slug: slug, Transaction: dto.TransactionDTO{AmountPaidCents: 1000}}
}

func profileWithGroups(groups ...dto.GroupType) *dto.ProfileDTO {
	return &dto.ProfileDTO{Groups: groups}
}

func TestGeneralPolicyRegularMemberTransitions(t *testing.T) {
	policy := NewGeneralPolicy()
	member := profileWithGroups(dto.GroupMember)

	tests := []struct {
		name         string
		current      *dto.MembershipDTO
		requested    string
		wantAllowed  bool
		wantPurchase dto.PurchaseType
		wantReason   dto.TierUnavailableReason
	}{
		{"no membership -> day", nil, "day", true, dto.PurchaseNew, ""},
		{"no membership -> basic", nil, "basic", true, dto.PurchaseNew, ""},
		{"no membership -> lounge", nil, "lounge", true, dto.PurchaseNew, ""},
		{"owns day -> day is already owned", membershipOf("day"), "day", false, "", dto.ReasonAlreadyOwned},
		{"owns day -> basic", membershipOf("day"), "basic", true, dto.PurchaseNew, ""},
		{"owns day -> lounge", membershipOf("day"), "lounge", true, dto.PurchaseNew, ""},
		{"owns basic -> basic is already owned", membershipOf("basic"), "basic", false, "", dto.ReasonAlreadyOwned},
		{"owns basic -> lounge is an upgrade", membershipOf("basic"), "lounge", true, dto.PurchaseUpgrade, ""},
		{"owns basic -> day is not eligible", membershipOf("basic"), "day", false, "", dto.ReasonNotEligibleCurrent},
		{"owns lounge -> lounge is already owned", membershipOf("lounge"), "lounge", false, "", dto.ReasonAlreadyOwned},
		{"owns lounge -> day is not eligible", membershipOf("lounge"), "day", false, "", dto.ReasonNotEligibleCurrent},
		{"owns lounge -> basic is not eligible", membershipOf("lounge"), "basic", false, "", dto.ReasonNotEligibleCurrent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := policy.Evaluate(member, tt.current, generalTier(tt.requested))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Allowed != tt.wantAllowed {
				t.Fatalf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if tt.wantAllowed && result.PurchaseType != tt.wantPurchase {
				t.Fatalf("PurchaseType = %q, want %q", result.PurchaseType, tt.wantPurchase)
			}
			if !tt.wantAllowed && result.Reason != tt.wantReason {
				t.Fatalf("Reason = %q, want %q", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestGeneralPolicyExecutivesBlockedFromGeneralTiers(t *testing.T) {
	policy := NewGeneralPolicy()

	for _, group := range []dto.GroupType{dto.GroupExecutive, dto.GroupDirector, dto.GroupBoard} {
		for _, slug := range []string{"day", "basic", "lounge"} {
			t.Run(string(group)+"_"+slug, func(t *testing.T) {
				result, err := policy.Evaluate(profileWithGroups(group), nil, generalTier(slug))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result.Allowed {
					t.Fatalf("expected %s to be blocked from %s", group, slug)
				}
				if result.Reason != dto.ReasonExecutiveRestricted {
					t.Fatalf("Reason = %q, want %q", result.Reason, dto.ReasonExecutiveRestricted)
				}
			})
		}
	}
}

func TestGeneralPolicyCompetitivePlayersBlockedFromGeneralTiers(t *testing.T) {
	policy := NewGeneralPolicy()

	for _, slug := range []string{"day", "basic", "lounge"} {
		t.Run(slug, func(t *testing.T) {
			result, err := policy.Evaluate(profileWithGroups(dto.GroupCompetitiveTeam), nil, generalTier(slug))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Allowed {
				t.Fatalf("expected competitive team to be blocked from %s", slug)
			}
			if result.Reason != dto.ReasonCompetitiveRestricted {
				t.Fatalf("Reason = %q, want %q", result.Reason, dto.ReasonCompetitiveRestricted)
			}
		})
	}
}

func TestGeneralPolicyExecAndCompetitivePrioritizesExecReason(t *testing.T) {
	policy := NewGeneralPolicy()
	profile := profileWithGroups(dto.GroupExecutive, dto.GroupCompetitiveTeam)

	result, err := policy.Evaluate(profile, nil, generalTier("basic"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected exec+competitive profile to be blocked from a general tier")
	}
	if result.Reason != dto.ReasonExecutiveRestricted {
		t.Fatalf("Reason = %q, want %q (exec should take priority)", result.Reason, dto.ReasonExecutiveRestricted)
	}
}

func TestGeneralPolicyExecutiveCanBuyTheirOwnTier(t *testing.T) {
	policy := NewGeneralPolicy()
	exec := profileWithGroups(dto.GroupExecutive)

	result, err := policy.Evaluate(exec, nil, generalTier("executive"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed || result.PurchaseType != dto.PurchaseNew {
		t.Fatalf("expected an exec with no membership to be allowed to buy the executive tier, got %+v", result)
	}
}

func TestGeneralPolicyExecutiveAlreadyOwningExecutiveTier(t *testing.T) {
	policy := NewGeneralPolicy()
	exec := profileWithGroups(dto.GroupExecutive)

	result, err := policy.Evaluate(exec, membershipOf("executive"), generalTier("executive"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected already-owned executive tier to be blocked")
	}
	if result.Reason != dto.ReasonAlreadyOwned {
		t.Fatalf("Reason = %q, want %q", result.Reason, dto.ReasonAlreadyOwned)
	}
}
