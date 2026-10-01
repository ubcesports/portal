package repository

import "go.uber.org/fx"

var Module = fx.Module("repository",
	fx.Provide(
		NewHealthRepository,
		NewProfileRepository,
		NewAdminRepository,
		NewExecProfileRepository,
		NewMembershipRepository,
	),
)
