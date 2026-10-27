package e2etest

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/thecodearcher/limen/plugins/oauth"
	oauthgeneric "github.com/thecodearcher/limen/plugins/oauth-generic"
)

func TestProviderWithLimenOAuthClient(t *testing.T) {
	callback := "http://localhost:18080/auth/oauth/zetrova/callback"
	p := NewProvider("", callback)
	server := httptest.NewServer(p.Handler())
	defer server.Close()
	p.BaseURL = server.URL
	provider := oauthgeneric.New(
		oauthgeneric.WithName("zetrova"), oauthgeneric.WithClientID("e2e"), oauthgeneric.WithClientSecret("e2e"),
		oauthgeneric.WithRedirectURL(callback), oauthgeneric.WithDiscoveryURL(server.URL+"/.well-known/openid-configuration"),
		oauthgeneric.WithMapUserInfo(func(raw map[string]any) (*oauth.ProviderUserInfo, error) {
			return &oauth.ProviderUserInfo{ID: raw["sub"].(string), Email: raw["email"].(string)}, nil
		}),
	)
	verifier := strings.Repeat("v", 43)
	hash := sha256.Sum256([]byte(verifier))
	form := url.Values{"persona": {"member"}, "redirect_uri": {callback}, "state": {"test-state"}, "code_challenge": {base64.RawURLEncoding.EncodeToString(hash[:])}}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.PostForm(server.URL+"/authorize", form)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 303 {
		t.Fatalf("authorization: %d", response.StatusCode)
	}
	redirect, err := url.Parse(response.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if redirect.Query().Get("state") != "test-state" {
		t.Fatal("OAuth state was lost")
	}
	code := redirect.Query().Get("code")
	token, err := provider.(oauth.TokenExchanger).ExchangeAuthorizationCode(context.Background(), code, verifier, callback)
	if err != nil {
		t.Fatal(err)
	}
	info, err := provider.GetUserInfo(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != "member" || info.Email != "member@e2e.invalid" {
		t.Fatalf("unexpected user: %+v", info)
	}
	if _, err := provider.(oauth.TokenExchanger).ExchangeAuthorizationCode(context.Background(), code, verifier, callback); err == nil {
		t.Fatal("authorization codes must be single-use")
	}
}

func TestProviderRejectsInvalidRedirectAndToken(t *testing.T) {
	p := NewProvider("http://localhost:18081", "http://localhost:18080/callback")
	for _, path := range []string{"/authorize?redirect_uri=https://example.com", "/userinfo"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		p.Handler().ServeHTTP(response, request)
		if response.Code < 400 {
			t.Fatalf("accepted invalid request: %s", path)
		}
	}
}

func TestLocalTransportBlocksExternalRequests(t *testing.T) {
	transport := LocalTransport{Base: http.DefaultTransport}
	req, _ := http.NewRequest(http.MethodGet, "https://api.stripe.com/v1/checkout/sessions", nil)
	if _, err := transport.RoundTrip(req); err == nil {
		t.Fatal("external request was not blocked")
	}
}

func TestMailIsCapturedLocally(t *testing.T) {
	p := NewPayments("", "", "unused")
	mux := http.NewServeMux()
	p.Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	client := &http.Client{Transport: LocalTransport{Base: http.DefaultTransport, MailURL: server.URL}}
	response, err := client.Post("https://api.resend.com/emails", "application/json", strings.NewReader(`{"to":["member@e2e.invalid"],"subject":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("mail status: %d", response.StatusCode)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.emails) != 1 || !json.Valid(p.emails[0]) {
		t.Fatalf("email not captured: %v", p.emails)
	}
}
