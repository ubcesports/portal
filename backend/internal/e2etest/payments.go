package e2etest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v86"
)

type checkout struct {
	Session    stripe.CheckoutSession
	SuccessURL string
	CancelURL  string
}

// Payments replaces the external Stripe HTTP boundary, not the application's
// checkout service or webhook handler. Amounts match the synthetic catalog.
type Payments struct {
	BaseURL     string
	APIURL      string
	Secret      string
	mu          sync.Mutex
	sessions    map[string]*checkout
	idempotency map[string]string
	emails      []json.RawMessage
}

func NewPayments(baseURL, apiURL, secret string) *Payments {
	return &Payments{BaseURL: baseURL, APIURL: apiURL, Secret: secret, sessions: map[string]*checkout{}, idempotency: map[string]string{}}
}

var checkoutPage = template.Must(template.New("checkout").Parse(`<!doctype html><html><body>
<h1>E2E checkout</h1><form method="post" action="/checkout/{{.}}/complete"><button>Pay</button></form>
<form method="post" action="/checkout/{{.}}/cancel"><button>Cancel payment</button></form></body></html>`))

func (p *Payments) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/checkout/sessions", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		key := r.Header.Get("Idempotency-Key")
		if id, ok := p.idempotency[key]; ok && key != "" {
			respond(w, p.sessions[id].Session)
			return
		}
		price := r.FormValue("line_items[0][price]")
		amount := int64(1000)
		if strings.HasSuffix(price, "_student") {
			amount = 500
		}
		if !strings.HasPrefix(price, "price_e2e_") {
			http.Error(w, "unsupported fixture price", http.StatusBadRequest)
			return
		}
		id := "cs_e2e_" + uuid.NewString()
		c := &checkout{Session: stripe.CheckoutSession{
			ID: id, Object: "checkout.session", URL: p.BaseURL + "/checkout/" + id,
			Status: stripe.CheckoutSessionStatusOpen, PaymentStatus: stripe.CheckoutSessionPaymentStatusUnpaid,
			AmountTotal: amount, Currency: stripe.CurrencyCAD,
			Metadata: map[string]string{"transaction_id": r.FormValue("metadata[transaction_id]"), "user_id": r.FormValue("metadata[user_id]")},
		}, SuccessURL: r.FormValue("success_url"), CancelURL: r.FormValue("cancel_url")}
		p.sessions[id] = c
		p.idempotency[key] = id
		respond(w, c.Session)
	})
	mux.HandleFunc("GET /v1/checkout/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		c, ok := p.sessions[r.PathValue("id")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		respond(w, c.Session)
	})
	mux.HandleFunc("POST /v1/checkout/sessions/{id}/expire", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		c, ok := p.sessions[r.PathValue("id")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if c.Session.Status == stripe.CheckoutSessionStatusComplete {
			http.Error(w, "already paid", 400)
			return
		}
		c.Session.Status = stripe.CheckoutSessionStatusExpired
		respond(w, c.Session)
	})
	mux.HandleFunc("GET /checkout/{id}", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		_, ok := p.sessions[r.PathValue("id")]
		p.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = checkoutPage.Execute(w, r.PathValue("id"))
	})
	mux.HandleFunc("POST /checkout/{id}/{action}", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		c, ok := p.sessions[r.PathValue("id")]
		if !ok {
			p.mu.Unlock()
			http.NotFound(w, r)
			return
		}
		eventType := "checkout.session.completed"
		destination := c.SuccessURL
		switch r.PathValue("action") {
		case "complete":
			c.Session.Status = stripe.CheckoutSessionStatusComplete
			c.Session.PaymentStatus = stripe.CheckoutSessionPaymentStatusPaid
			c.Session.PaymentIntent = &stripe.PaymentIntent{ID: "pi_" + c.Session.ID}
		case "cancel":
			if c.Session.Status == stripe.CheckoutSessionStatusComplete {
				p.mu.Unlock()
				http.Error(w, "already paid", 400)
				return
			}
			c.Session.Status = stripe.CheckoutSessionStatusExpired
			eventType = "checkout.session.expired"
			destination = c.CancelURL
		default:
			p.mu.Unlock()
			http.NotFound(w, r)
			return
		}
		session := c.Session
		p.mu.Unlock()
		// Completing again deliberately delivers another event for the same session,
		// allowing the suite to exercise the real application's idempotency.
		if err := p.deliver(r, eventType, session); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, destination, http.StatusSeeOther)
	})
	mux.HandleFunc("POST /emails", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil || !json.Valid(body) {
			http.Error(w, "invalid email", 400)
			return
		}
		p.mu.Lock()
		p.emails = append(p.emails, json.RawMessage(body))
		p.mu.Unlock()
		respond(w, map[string]string{"id": uuid.NewString()})
	})
	mux.HandleFunc("GET /emails", func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()
		respond(w, p.emails)
	})
}

func (p *Payments) deliver(r *http.Request, eventType string, session stripe.CheckoutSession) error {
	now := time.Now().Unix()
	payload, err := json.Marshal(map[string]any{"id": "evt_" + uuid.NewString(), "object": "event", "api_version": stripe.APIVersion, "created": now, "type": eventType, "data": map[string]any{"object": session}})
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(p.Secret))
	fmt.Fprintf(mac, "%d.", now)
	_, _ = mac.Write(payload)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, p.APIURL+"/webhooks/stripe", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%s", now, hex.EncodeToString(mac.Sum(nil))))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("webhook failed: %s: %s", response.Status, body)
	}
	return nil
}
