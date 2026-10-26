package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type ExecProfileHandler struct {
	execProfileService *service.ExecProfileService
}

func NewExecProfileHandler(execProfileService *service.ExecProfileService) *ExecProfileHandler {
	return &ExecProfileHandler{execProfileService: execProfileService}
}

/*
Returns a list of all executive profiles, grouped by their display group.

API URL: GET /exec-profiles

Args: none

Returns:

	execs: a list of executive profiles, grouped by their display group ("executive", "central_director", "game_director", "board", "president")
		   (HTTP 200).

Raises:

	500: unable to load executive profiles
*/
func (h *ExecProfileHandler) GetExecProfiles(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	execProfiles, err := h.execProfileService.GetExecProfiles(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load executive profiles",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load executive profiles", requestId)
		return
	}

	groupedProfiles := groupExecProfilesByDisplayGroup(execProfiles)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"execs": groupedProfiles,
	})
}

/*
Returns the executive profile for the currently authenticated user.

API URL: GET /exec-profile

Returns:

	exec_profile: the current user's executive profile (HTTP 200).

Raises:

	401: unauthorized user
	404: executive profile not found
	500: unable to load executive profile
*/
func (h *ExecProfileHandler) GetCurrentExecProfile(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	execProfile, err := h.execProfileService.GetExecProfileByUserID(r.Context(), userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", "Executive profile not found", requestId)
			return
		}

		slog.ErrorContext(r.Context(), "unable to load executive profile",
			"error", err,
			"request_id", requestId,
			"user_id", userId,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load executive profile", requestId)
		return
	}

	util.WriteJson(w, http.StatusOK, map[string]*dto.ExecProfileDTO{"exec_profile": execProfile})
}

/*
Updates an existing social link in the executive profile of the currently authenticated user, or adds one if it doesn't exist.

API URL: PUT /exec-profile/social-links

Args (query params):

	platform: the social media platform (e.g., "instagram", "linkedin") for the link.
	url: the URL of the social media profile to be added.

Returns:

	Success response (HTTP 200) and message

Raises:

	400: invalid request parameters
	401: unauthorized user
	500: unable to update social link
*/
func (h *ExecProfileHandler) UpdateExecSocialLink(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	var request dto.ExecProfileSocialLinkDTO
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	if request.Platform == "" || request.URL == "" {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Missing required fields: platform and url are required", requestId)
		return
	}

	err := h.execProfileService.UpdateExecSocialLink(r.Context(), userId, db.ExecSocialPlatformType(request.Platform), request.URL)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to update social link",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to update social link", requestId)
		return
	}

	util.WriteApiResponse(w, http.StatusOK, "SUCCESS", "Social link updated successfully", requestId)
}

/*
Deletes a social link from the executive profile of the currently authenticated user.

API URL: DELETE /exec-profile/social-links

Args (query params):

	platform: the social media platform (e.g., "instagram", "linkedin") for the link.

Returns:

	Success response (HTTP 200) and message

Raises:

	400: invalid request parameters
	401: unauthorized user
	500: unable to delete social link
*/
func (h *ExecProfileHandler) DeleteExecSocialLink(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	var request struct {
		Platform db.ExecSocialPlatformType `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	if request.Platform == "" {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Missing required query parameters: platform is required", requestId)
		return
	}

	err := h.execProfileService.DeleteExecSocialLink(r.Context(), userId, db.ExecSocialPlatformType(request.Platform))
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to delete social link",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to delete social link", requestId)
		return
	}

	util.WriteApiResponse(w, http.StatusOK, "SUCCESS", "Social link deleted successfully", requestId)
}

/*
Updates the title of the executive profile of the currently authenticated user.

API URL: PATCH /exec-profile/title

Args (query params):

	title: the new title for the executive profile.

Returns:

	Success response (HTTP 200) and message

Raises:

	400: invalid request parameters
	401: unauthorized user
	500: unable to update title
*/
func (h *ExecProfileHandler) UpdateExecProfileTitle(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	var request struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	if request.Title == "" {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Missing required field: title is required", requestId)
		return
	}

	_, err := h.execProfileService.UpdateExecProfileTitle(r.Context(), userId, request.Title)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to update exec profile title",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to update exec profile title", requestId)
		return
	}

	util.WriteApiResponse(w, http.StatusOK, "SUCCESS", "Exec profile title updated successfully", requestId)
}

/*
	Private functions
*/

func groupExecProfilesByDisplayGroup(execProfiles []*dto.ExecProfileDTO) map[dto.ExecDisplayGroupType][]*dto.ExecProfileDTO {
	groupedProfiles := make(map[dto.ExecDisplayGroupType][]*dto.ExecProfileDTO)
	for _, profile := range execProfiles {
		groupedProfiles[profile.DisplayGroup] = append(groupedProfiles[profile.DisplayGroup], profile)
	}
	return groupedProfiles
}
