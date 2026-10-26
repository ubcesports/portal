// E2E-only composition. Production images build cmd/api, never this command.
package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v86"
	"github.com/thecodearcher/limen"
	"github.com/ubcesports/memberships/internal/auth"
	"github.com/ubcesports/memberships/internal/database"
	"github.com/ubcesports/memberships/internal/e2etest"
	"github.com/ubcesports/memberships/internal/handlers"
	"github.com/ubcesports/memberships/internal/logging"
	"github.com/ubcesports/memberships/internal/mailer"
	"github.com/ubcesports/memberships/internal/membershippolicy"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/server"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/stripeclient"
	"go.uber.org/fx"
	_ "time/tzdata"
)

func main() {
	// Never load developer .env files, allow a remote DB, or start the scheduler.
	if os.Getenv("POSTGRES_DB") != "portal_e2e" || os.Getenv("POSTGRES_HOST") != "127.0.0.1" || os.Getenv("BASE_URL") != "http://localhost:18080" {
		log.Fatal("E2E entrypoint requires the isolated local test environment; run make e2e")
	}
	http.DefaultTransport = e2etest.LocalTransport{Base: http.DefaultTransport, MailURL: "http://127.0.0.1:18081"}
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL: stripe.String("http://127.0.0.1:18081"), HTTPClient: &http.Client{Timeout: 10 * time.Second},
		MaxNetworkRetries: stripe.Int64(0), EnableTelemetry: stripe.Bool(false),
	}))
	provider := e2etest.NewProvider("http://localhost:18081", os.Getenv("OAUTH_CALLBACK_URL"))
	payments := e2etest.NewPayments(provider.BaseURL, os.Getenv("BASE_URL"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	mux := http.NewServeMux()
	mux.Handle("/", provider.Handler())
	payments.Register(mux)
	listener, err := net.Listen("tcp", "127.0.0.1:18081")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	defer srv.Close()
	logging.Configure()
	fx.New(
		database.Module,
		fx.Provide(func(db *sql.DB) (*limen.Limen, error) {
			// All test identities share one loopback IP. Do not let the process-wide
			// auth rate limiter leak request counts across database resets.
			// Production cmd/api retains Limen's default rate limiting.
			return auth.NewLimenWithDiscoveryURL(db, provider.BaseURL+"/.well-known/openid-configuration",
				limen.WithHTTPRateLimiter(limen.WithRateLimiterEnabled(false)),
			)
		}),
		repository.Module, service.Module, handlers.Module, server.Module,
		stripeclient.Module, membershippolicy.Module, mailer.Module,
	).Run()
}
