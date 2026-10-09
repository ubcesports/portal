package service

import (
	"errors"
	"testing"

	"github.com/ubcesports/memberships/internal/dto"
)

func TestNormalizeMembershipInvitation(t *testing.T) {
	request, err := normalizeMembershipInvitation(dto.CreateMembershipInvitationRequest{
		Email:           " Member@Example.com ",
		TierID:          "2d746a56-c977-49e0-a04c-20504cdb07c0",
		AmountPaidCents: 2500,
		PaymentMethod:   dto.PaymentMethodCash,
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Email != "member@example.com" {
		t.Fatalf("unexpected normalized email: %#v", request)
	}
}

func TestNormalizeMembershipInvitationRejectsInvalidValues(t *testing.T) {
	tests := []dto.CreateMembershipInvitationRequest{
		{Email: "not-an-email", TierID: "2d746a56-c977-49e0-a04c-20504cdb07c0", PaymentMethod: dto.PaymentMethodCash},
		{Email: "ada@example.com", TierID: "2d746a56-c977-49e0-a04c-20504cdb07c0", AmountPaidCents: -1, PaymentMethod: dto.PaymentMethodCash},
	}
	for _, request := range tests {
		if _, err := normalizeMembershipInvitation(request); !errors.Is(err, ErrInvalidMembershipInvitation) {
			t.Fatalf("expected invalid purchase error for %#v, got %v", request, err)
		}
	}
}

func TestNormalizeMembershipInvitationRejectsInvalidTierID(t *testing.T) {
	_, err := normalizeMembershipInvitation(dto.CreateMembershipInvitationRequest{
		Email:         "ada@example.com",
		TierID:        "invalid",
		PaymentMethod: dto.PaymentMethodCash,
	})
	if !errors.Is(err, ErrInvalidMembershipInvitationTier) {
		t.Fatalf("expected invalid tier error, got %v", err)
	}
}

func TestNormalizeMembershipInvitationRejectsOnlinePaymentMethod(t *testing.T) {
	_, err := normalizeMembershipInvitation(dto.CreateMembershipInvitationRequest{
		Email:         "ada@example.com",
		TierID:        "2d746a56-c977-49e0-a04c-20504cdb07c0",
		PaymentMethod: dto.PaymentMethodStripe,
	})
	if !errors.Is(err, ErrOfflinePaymentMethod) {
		t.Fatalf("expected offline payment method error, got %v", err)
	}
}
