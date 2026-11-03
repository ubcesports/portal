package e2etest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
)

func TestCheckoutUsesStripeClientAndSignedWebhook(t *testing.T) {
	const secret = "whsec_e2e"
	events := make(chan stripe.Event, 2)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			http.Error(w, "read", 400)
			return
		}
		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), secret)
		if err != nil {
			t.Error(err)
			http.Error(w, "signature", 400)
			return
		}
		events <- event
		w.WriteHeader(200)
	}))
	defer api.Close()
	p := NewPayments("", api.URL, secret)
	mux := http.NewServeMux()
	p.Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	p.BaseURL = server.URL
	client := stripe.NewClient("sk_test_e2e", stripe.WithBackends(stripe.NewBackendsWithConfig(&stripe.BackendConfig{
		URL: stripe.String(server.URL), HTTPClient: server.Client(), MaxNetworkRetries: stripe.Int64(0), EnableTelemetry: stripe.Bool(false),
	})))
	params := &stripe.CheckoutSessionCreateParams{
		Mode: stripe.String("payment"), SuccessURL: stripe.String("http://localhost:13000/checkout?success=true"), CancelURL: stripe.String("http://localhost:13000/checkout?canceled=true"),
		Metadata:  map[string]string{"transaction_id": "transaction", "user_id": "user"},
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{{Price: stripe.String("price_e2e_day_community"), Quantity: stripe.Int64(1)}},
	}
	params.SetIdempotencyKey("transaction")
	session, err := client.V1CheckoutSessions.Create(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if session.AmountTotal != 1000 || session.Metadata["transaction_id"] != "transaction" {
		t.Fatalf("invalid session: %+v", session)
	}
	again, err := client.V1CheckoutSessions.Create(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != session.ID {
		t.Fatal("idempotency key ignored")
	}
	httpClient := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for i := 0; i < 2; i++ {
		response, err := httpClient.Post(session.URL+"/complete", "application/x-www-form-urlencoded", nil)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 303 {
			t.Fatalf("completion failed: %d", response.StatusCode)
		}
		event := <-events
		if event.Type != stripe.EventTypeCheckoutSessionCompleted {
			t.Fatalf("wrong event: %s", event.Type)
		}
	}
	retrieved, err := client.V1CheckoutSessions.Retrieve(context.Background(), session.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		t.Fatal("payment status not persisted")
	}
}
