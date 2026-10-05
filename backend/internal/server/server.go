package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thecodearcher/limen"
	"github.com/ubcesports/memberships/internal/auth"
	"github.com/ubcesports/memberships/internal/handlers"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(provideRouter),
	fx.Invoke(startServer),
)

type RouterParams struct {
	fx.In

	HealthHandler        *handlers.HealthHandler
	ProfileHandler       *handlers.ProfileHandler
	AdminHandler         *handlers.AdminHandler
	MembershipHandler    *handlers.MembershipHandler
	StripeWebhookHandler *handlers.StripeWebhookHandler
	AnalyticsHandler     *handlers.AnalyticsHandler
	Limen                *limen.Limen
}

// provideRouter builds the HTTP router with shared middleware and public,
// authenticated, onboarded, and admin-only route groups.
func provideRouter(params RouterParams) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(logger)
	r.Use(recoverer)
	r.Use(corsMiddleware)

	r.Mount("/", params.Limen.Handler())

	// All public routes
	r.Get("/health", params.HealthHandler.IsDatabaseHealthy)
	r.Get("/membership/tiers", params.MembershipHandler.GetPublicTiersWithPrices)
	r.Post("/webhooks/stripe", params.StripeWebhookHandler.Handle)

	// All protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(params.Limen))

		r.Get("/profile", params.ProfileHandler.GetCurrentProfile)

		r.Post("/onboard", params.ProfileHandler.OnboardUser)
	})

	// All onboarded routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(params.Limen))
		r.Use(auth.RequireOnboarded)

		r.Get("/onboard/check", params.ProfileHandler.GetIsUserOnboarded)

		r.Get("/membership/me/current", params.MembershipHandler.GetCurrentMembershipWithTransaction)
		r.Get("/membership/me/all", params.MembershipHandler.GetAllMembershipsWithTransactions)
		r.Get("/membership/tiers/eligible", params.MembershipHandler.GetEligibleTiersWithPrices)
		r.Post("/membership/checkout", params.MembershipHandler.CreateMembershipCheckoutSession)
	})

	// All admin routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(params.Limen))
		r.Use(auth.RequireRole("admin"))

		r.Get("/admin/users", params.AdminHandler.GetUsers)
		r.Get("/admin/users/export", params.AdminHandler.ExportUsersCSV)
		r.Get("/admin/membership-tiers", params.AdminHandler.GetMembershipTierOptions)
		r.Get("/admin/users/{id}", params.AdminHandler.GetUser)
		r.Get("/admin/users/{id}/memberships", params.AdminHandler.GetUserMemberships)
		r.Patch("/admin/users/{id}", params.AdminHandler.UpdateUser)
		r.Get("/admin/audit-logs", params.AdminHandler.GetAdminAuditLogs)
		r.Get("/admin/audit-logs/export", params.AdminHandler.ExportAuditLogsCSV)
		r.Post("/admin/membership/add/{id}", params.AdminHandler.AddMembershipToUser)
		r.Get("/admin/memberships/eligible/{id}", params.AdminHandler.GetEligibleTiersWithPricesById)

		r.Get("/admin/analytics/summary", params.AnalyticsHandler.GetSummary)
		r.Get("/admin/analytics/memberships-over-time", params.AnalyticsHandler.GetMembershipsBoughtOverTime)
		r.Get("/admin/analytics/revenue-over-time", params.AnalyticsHandler.GetRevenueOverTime)
		r.Get("/admin/analytics/export", params.AnalyticsHandler.ExportCSV)
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range auth.TrustedFrontendOrigins() {
		allowedOrigins[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowedOrigins[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startServer(lc fx.Lifecycle, r *chi.Mux) error {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return errors.New("BASE_URL environment variable is required")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("error parsing BASE_URL: %w", err)
	}

	srv := &http.Server{Addr: parsed.Host, Handler: r}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			slog.InfoContext(ctx, "server running", "base_url", baseURL)
			go func() {
				if err := srv.ListenAndServe(); err != http.ErrServerClosed {
					slog.Error("server stopped unexpectedly", "error", err)
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.InfoContext(ctx, "stopping server...")
			if err := srv.Shutdown(ctx); err != nil {
				return fmt.Errorf("error shutting down server: %w", err)
			}
			return nil
		},
	})
	return nil
}
