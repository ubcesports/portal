package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type ProfileHandler struct {
	profileService *service.ProfileService
	roleService    *service.RoleService
}

func NewProfileHandler(profileService *service.ProfileService, roleService *service.RoleService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService, roleService: roleService}
}

/*
Get the currently authenticated user's profile.

API URL: GET /profile

Args:

	None

Returns:

	response body containing the current user's profile under the "user" key (HTTP 200).
	The profile includes "roles": the custom roles assigned to the user, each with
	its id, name and permission_keys.

Raises:

	401: unauthorized user
	500: unable to load profile
*/
func (h *ProfileHandler) GetCurrentProfile(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	profile, err := h.profileService.GetProfileByUserID(r.Context(), userId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load profile",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
			"user_id", userId,
		)

		util.WriteApiResponse(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Unable to load profile",
			requestId,
		)
		return
	}

	roles, err := h.roleService.GetForUser(r.Context(), userId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load user roles",
			"error", err,
			"request_id", requestId,
			"user_id", userId,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load profile", requestId)
		return
	}
	if roles == nil {
		roles = []dto.UserRolePermissions{}
	}
	profile.Roles = roles

	util.WriteJson(w, http.StatusOK, map[string]dto.ProfileDTO{"user": *profile})
}

/*
Check whether the currently authenticated user is onboarded.
Only returns HTTP 200, as middleware auto-handles non-onboarded users.

API URL: GET /onboard/check

Args:

	None

Returns:

	None (HTTP 200)

Raises:

	None
*/
func (h *ProfileHandler) GetIsUserOnboarded(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

/*
Onboard the currently authenticated user.

API URL: POST /onboard

Args:

	request body: dto.OnboardUserRequest

Returns:

	response body with success code and message (HTTP 200)

Raises:

	400: invalid request body or validation error
	401: unauthorized user
	409: conflict, when the user is already onboarded
	500: internal error, when onboarding fails unexpectedly
*/
func (h *ProfileHandler) OnboardUser(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	// Ensure request body is valid
	var onboardUserRequest dto.OnboardUserRequest
	if err := json.NewDecoder(r.Body).Decode(&onboardUserRequest); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	// Onboard user
	if err := h.profileService.OnboardUser(r.Context(), userId, onboardUserRequest); err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			util.WriteApiResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestId)

		case errors.Is(err, service.ErrConflict):
			util.WriteApiResponse(w, http.StatusConflict, "CONFLICT", err.Error(), requestId)

		default:
			slog.ErrorContext(r.Context(), "unable to onboard user",
				"error", err,
				"request_id", middleware.GetReqID(r.Context()),
				"user_id", userId,
			)
			util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error onboarding user. Please try again.", requestId)
		}

		return
	}

	util.WriteApiResponse(w, http.StatusOK, "OK", "User onboarded successfully!", requestId)
}
