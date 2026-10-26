// Package e2etest contains local-only infrastructure used by cmd/e2e.
// It is deliberately not imported by the production entrypoint.
package e2etest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type grant struct {
	Persona   string
	Redirect  string
	Challenge string
	Expires   time.Time
}

// Provider implements only the authorization-code/userinfo contract used by
// Limen. This is not a production identity provider.
type Provider struct {
	BaseURL     string
	CallbackURL string
	mu          sync.Mutex
	codes       map[string]grant
	tokens      map[string]grant
}

func NewProvider(baseURL, callbackURL string) *Provider {
	return &Provider{BaseURL: baseURL, CallbackURL: callbackURL, codes: map[string]grant{}, tokens: map[string]grant{}}
}

var loginPage = template.Must(template.New("login").Parse(`<!doctype html><html><body>
<h1>E2E identity provider</h1>
<form method="post">
<input type="hidden" name="state" value="{{.State}}">
<input type="hidden" name="redirect_uri" value="{{.Redirect}}">
<input type="hidden" name="code_challenge" value="{{.Challenge}}">
<label for="persona">Test identity</label>
<select id="persona" name="persona">
<option value="member">Member</option><option value="admin">Administrator</option>
<option value="new">New member</option><option value="history">Member with history</option>
</select><button type="submit">Sign in</button></form></body></html>`))

func (p *Provider) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]string{"issuer": p.BaseURL, "authorization_endpoint": p.BaseURL + "/authorize", "token_endpoint": p.BaseURL + "/token", "userinfo_endpoint": p.BaseURL + "/userinfo"})
	})
	mux.HandleFunc("GET /authorize", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("redirect_uri") != p.CallbackURL || r.URL.Query().Get("code_challenge_method") != "S256" {
			http.Error(w, "invalid authorization request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginPage.Execute(w, map[string]string{"State": r.URL.Query().Get("state"), "Redirect": p.CallbackURL, "Challenge": r.URL.Query().Get("code_challenge")})
	})
	mux.HandleFunc("POST /authorize", func(w http.ResponseWriter, r *http.Request) {
		persona := r.FormValue("persona")
		if !validPersona(persona) || r.FormValue("redirect_uri") != p.CallbackURL || r.FormValue("state") == "" {
			http.Error(w, "invalid authorization request", http.StatusBadRequest)
			return
		}
		code := uuid.NewString()
		p.mu.Lock()
		p.codes[code] = grant{Persona: persona, Redirect: p.CallbackURL, Challenge: r.FormValue("code_challenge"), Expires: time.Now().Add(time.Minute)}
		p.mu.Unlock()
		target, _ := url.Parse(p.CallbackURL)
		q := target.Query()
		q.Set("code", code)
		q.Set("state", r.FormValue("state"))
		target.RawQuery = q.Encode()
		http.Redirect(w, r, target.String(), http.StatusSeeOther)
	})
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		code := r.FormValue("code")
		g, ok := p.codes[code]
		delete(p.codes, code)
		hash := sha256.Sum256([]byte(r.FormValue("code_verifier")))
		if !ok || time.Now().After(g.Expires) || g.Redirect != r.FormValue("redirect_uri") || g.Challenge != base64.RawURLEncoding.EncodeToString(hash[:]) {
			http.Error(w, "invalid grant", http.StatusBadRequest)
			return
		}
		token := uuid.NewString()
		g.Expires = time.Now().Add(time.Hour)
		p.tokens[token] = g
		respond(w, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 3600, "scope": "openid profile email"})
	})
	mux.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		g, ok := p.tokens[strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")]
		p.mu.Unlock()
		if !ok || time.Now().After(g.Expires) {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		respond(w, map[string]any{"sub": g.Persona, "email": g.Persona + "@e2e.invalid", "name": "E2E " + g.Persona, "email_verified": true})
	})
	return mux
}

func validPersona(s string) bool {
	return s == "member" || s == "admin" || s == "new" || s == "history"
}
func respond(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

// LocalTransport prevents accidental real provider calls in the E2E process.
// Resend requests are redirected to the local capturing mail sink.
type LocalTransport struct {
	Base    http.RoundTripper
	MailURL string
}

func (t LocalTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	host := r.URL.Hostname()
	if host == "api.resend.com" && t.MailURL != "" {
		target, err := url.Parse(t.MailURL)
		if err != nil {
			return nil, err
		}
		clone := r.Clone(r.Context())
		clone.URL.Scheme = target.Scheme
		clone.URL.Host = target.Host
		clone.Host = target.Host
		return t.Base.RoundTrip(clone)
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, fmt.Errorf("E2E blocked external HTTP request to %s", host)
	}
	return t.Base.RoundTrip(r)
}
