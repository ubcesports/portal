package handlers

import "go.uber.org/fx"

var Module = fx.Module("handler",
	fx.Provide(
		NewHealthHandler,
		NewProfileHandler,
		NewAdminHandler,
		NewExecProfileHandler,
		NewMembershipHandler,
		NewStripeWebhookHandler,
	),
)
