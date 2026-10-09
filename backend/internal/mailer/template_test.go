package mailer

import (
	"strings"
	"testing"
)

func TestRenderEmailSupportsMembershipInvitation(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://app.example.com")

	html, err := RenderEmail(EmailData{
		Title:      "Redeem your UBCEA membership",
		Heading:    "Welcome to UBCEA!",
		Subheading: "Thanks for purchasing a membership in person.",
		Rows:       NewRows("Email", "ada@example.com", "Membership", "Lounge Tier"),
		CTAText:    "Sign up to redeem",
		CTAURL:     "https://app.example.com/login",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"Welcome to UBCEA!",
		"ada@example.com",
		"Lounge Tier",
		"https://app.example.com/login",
		"Sign up to redeem",
		"https://app.example.com/ubcea_logo.jpg",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("rendered invitation does not contain %q", expected)
		}
	}
}

func TestRenderEmailEscapesMembershipInvitationValues(t *testing.T) {
	html, err := RenderEmail(EmailData{
		Title:   "Invitation",
		Heading: "Welcome",
		Rows:    NewRows("Email", "member@example.com", "Membership", `<strong>tier</strong>`),
		CTAURL:  "https://example.com/login",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "<strong>") {
		t.Fatal("rendered invitation contains unescaped purchaser-controlled HTML")
	}
}
