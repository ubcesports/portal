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
	"github.com/ubcesports/memberships/internal/repository"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(provideRouter),
	fx.Invoke(startServer),
)

type RouterParams struct {
	fx.In

	HealthHandler               *handlers.HealthHandler
	ProfileHandler              *handlers.ProfileHandler
	AdminHandler                *handlers.AdminHandler
	ExecProfileHandler          *handlers.ExecProfileHandler
	MembershipHandler           *handlers.MembershipHandler
	MembershipInvitationHandler *handlers.MembershipInvitationHandler
	StripeWebhookHandler        *handlers.StripeWebhookHandler
	AnalyticsHandler            *handlers.AnalyticsHandler
	LoungeHandler               *handlers.LoungeHandler
	RoleHandler                 *handlers.RoleHandler
	AdminRepository             *repository.AdminRepository
	RoleRepository              *repository.RoleRepository
	Limen                       *limen.Limen
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
	r.Get("/exec-profiles", params.ExecProfileHandler.GetExecProfiles)
	r.Get("/lounge/stations", params.LoungeHandler.GetStations)

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

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(params.Limen))
		can := func(permission string) func(http.Handler) http.Handler {
			return auth.RequirePermission(params.RoleRepository, permission)
		}
		adminOnly := auth.RequireSystem(auth.SystemRoleAdmin)

		// Users
		r.With(can(auth.PermissionUsersView)).Get("/admin/users", params.AdminHandler.GetUsers)
		r.With(can(auth.PermissionUsersExport)).Get("/admin/users/export", params.AdminHandler.ExportUsersCSV)
		r.With(can(auth.PermissionUsersView)).Get("/admin/users/{id}", params.AdminHandler.GetUser)
		// Role changes and membership cancellations need further checks in the handler.
		r.With(can(auth.PermissionUsersUpdate)).Patch("/admin/users/{id}", params.AdminHandler.UpdateUser)

		// Membership tiers
		r.With(can(auth.PermissionMembershipTiersView)).Get("/admin/membership-tiers", params.AdminHandler.GetMembershipTierOptions)

		// Memberships
		r.With(can(auth.PermissionMembershipsView)).Get("/admin/users/{id}/memberships", params.AdminHandler.GetUserMemberships)
		r.With(can(auth.PermissionMembershipsCreate)).Get("/admin/memberships/eligible/{id}", params.AdminHandler.GetEligibleTiersWithPricesById)
		r.With(can(auth.PermissionMembershipsCreate)).Post("/admin/membership/add/{id}", params.AdminHandler.AddMembershipToUser)

		// Membership invitations
		r.With(can(auth.PermissionMembershipInvitationsView)).Get("/admin/membership-invitations", params.MembershipInvitationHandler.GetAll)
		r.With(can(auth.PermissionMembershipInvitationsView)).Get("/admin/membership-invitations/{id}", params.MembershipInvitationHandler.GetByID)
		r.With(can(auth.PermissionMembershipInvitationsCreate)).Post("/admin/membership-invitations", params.MembershipInvitationHandler.Create)
		r.With(can(auth.PermissionMembershipInvitationsDelete)).Delete("/admin/membership-invitations/{id}", params.MembershipInvitationHandler.Delete)

		// Exec profiles
		r.With(can(auth.PermissionExecProfilesUpdate)).Get("/admin/exec-profile/{id}", params.AdminHandler.GetExecProfile)
		r.With(can(auth.PermissionExecProfilesUpdate)).Patch("/admin/exec-profile/{id}", params.AdminHandler.UpdateExecProfile)

		// Audit logs
		r.With(can(auth.PermissionAuditLogsView)).Get("/admin/audit-logs", params.AdminHandler.GetAdminAuditLogs)
		r.With(can(auth.PermissionAuditLogsExport)).Get("/admin/audit-logs/export", params.AdminHandler.ExportAuditLogsCSV)

		// Analytics
		r.With(can(auth.PermissionAnalyticsView)).Get("/admin/analytics/summary", params.AnalyticsHandler.GetSummary)
		r.With(can(auth.PermissionAnalyticsView)).Get("/admin/analytics/memberships-over-time", params.AnalyticsHandler.GetMembershipsBoughtOverTime)
		r.With(can(auth.PermissionAnalyticsView)).Get("/admin/analytics/revenue-over-time", params.AnalyticsHandler.GetRevenueOverTime)
		r.With(can(auth.PermissionAnalyticsExport)).Get("/admin/analytics/export", params.AnalyticsHandler.ExportCSV)

		// Lounge
		r.With(can(auth.PermissionLoungeStationsUpdate)).Patch("/lounge/stations/{id}", params.LoungeHandler.UpdateStation)
		r.With(can(auth.PermissionLoungeSessionsCheckout)).Post("/lounge/sessions/{session_id}/check-out", params.LoungeHandler.CheckoutSession)

		// Roles
		// Managing roles can grant access the caller does not have, so it is
		// limited to system admins. Members see their own roles on GET /profile.
		r.With(adminOnly).Get("/admin/roles", params.RoleHandler.GetAll)
		r.With(adminOnly).Get("/admin/roles/{id}", params.RoleHandler.GetByID)
		r.With(adminOnly).Post("/admin/roles", params.RoleHandler.Create)
		r.With(adminOnly).Patch("/admin/roles/{id}", params.RoleHandler.Update)
		r.With(adminOnly).Delete("/admin/roles/{id}", params.RoleHandler.Delete)
		r.With(adminOnly).Post("/admin/roles/{id}/users", params.RoleHandler.AssignUser)
		r.With(adminOnly).Delete("/admin/roles/{id}/users/{userId}", params.RoleHandler.UnassignUser)
	})

	// All exec profile routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(params.Limen))
		r.Use(auth.RequireExecGroup(params.AdminRepository))

		r.Get("/exec-profile", params.ExecProfileHandler.GetCurrentExecProfile)
		r.Put("/exec-profile/social-links", params.ExecProfileHandler.UpdateExecSocialLink)
		r.Delete("/exec-profile/social-links", params.ExecProfileHandler.DeleteExecSocialLink)
		r.Patch("/exec-profile/title", params.ExecProfileHandler.UpdateExecProfileTitle)
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
