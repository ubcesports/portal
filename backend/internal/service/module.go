package service

import (
	"github.com/ubcesports/memberships/internal/membershippolicy"
	"go.uber.org/fx"
)

func provideProfileReader(
	profileService *ProfileService,
) membershippolicy.ProfileReader {
	return profileService
}

var Module = fx.Module("service",
	fx.Provide(
		NewHealthService,
		NewProfileService,
		NewAdminService,
		NewExecProfileService,
		NewMembershipService,

		provideProfileReader,
	),
)
