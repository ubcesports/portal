package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type LoungeHandler struct {
	loungeService *service.LoungeService
	adminService  *service.AdminService
}

func NewLoungeHandler(loungeService *service.LoungeService, adminService *service.AdminService) *LoungeHandler {
	return &LoungeHandler{loungeService: loungeService, adminService: adminService}
}

/*
Checks a member into a lounge station.

API URL: POST /lounge/sessions

Args (JSON body):

	student_id: member's student ID
	station_id: lounge station UUID
	game: optional game name

Returns:

	the created lounge session (HTTP 201)

Raises:

	400: invalid request
	401: user is not authenticated
	403: member does not have lounge access
	404: lounge station does not exist
	409: member already has an active session, station is closed, or station is in use
	500: lounge session could not be created
*/
func (h *LoungeHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	checkedInBy, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var request dto.CreateLoungeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body.", requestID)
		return
	}
	if strings.TrimSpace(request.StudentID) == "" {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_LOUNGE_SESSION", "Student ID is required.", requestID)
		return
	}
	if _, err := util.GetValidatedUUID(request.StationID); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_LOUNGE_SESSION", "Station ID must be a valid UUID.", requestID)
		return
	}

	session, err := h.loungeService.CreateLoungeSession(
		r.Context(),
		request.StudentID,
		request.StationID,
		request.Game,
		checkedInBy,
	)
	if err != nil {
		h.writeCreateSessionError(w, r, err)
		return
	}

	util.WriteJson(w, http.StatusCreated, session)
}

/*
Returns a list of all lounge stations, including layout coordinates and active session info.

API URL: GET /lounge/stations

Returns:

	[]dto.LoungeStationDTO (HTTP 200)

Raises:

	500: unable to load lounge stations
*/
func (h *LoungeHandler) GetStations(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	stations, err := h.loungeService.GetLoungeStations(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load lounge stations",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load lounge stations.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, stations)
}

/*
Updates a lounge station's label or closed status.

API URL: PATCH /lounge/stations/{id}

Args (JSON body):

	label: optional new station label
	closed: optional new closed status (boolean)

Returns:

	the updated lounge station (HTTP 200)

Raises:

	400: invalid UUID or request body
	404: lounge station not found
	500: unable to update lounge station
*/
func (h *LoungeHandler) UpdateStation(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	stationID := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(stationID); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid station ID.", requestID)
		return
	}

	var req dto.UpdateLoungeStationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body.", requestID)
		return
	}

	updated, err := h.loungeService.UpdateLoungeStation(r.Context(), stationID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", "Station not found.", requestID)
			return
		}
		slog.ErrorContext(r.Context(), "unable to update lounge station",
			"error", err,
			"station_id", stationID,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to update lounge station.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, updated)
}

/*
Returns the user matching the provided filter.

API URL: GET /lounge/members/lookup

Args (query params):

	exactly ONE of the following filters must be provided:

	full_name: optional case-insensitive name substring
	student_id: optional case-insensitive student ID substring
	email: optional case-insensitive email substring

Returns:

	user: user profile (HTTP 200)

Raises:

	400: invalid filter or pagination value
	401: user is not authenticated
	403: user is not an admin
	500: users could not be loaded
*/
func (h *LoungeHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	filters, err := parseLoungeLookupFilters(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, _, err := h.adminService.GetUsers(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load users",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		http.Error(w, "unable to load users", http.StatusInternalServerError)
		return
	}

	members := make([]dto.LoungeMemberLookupDTO, 0, len(users))
	for _, user := range users {
		members = append(members, dto.LoungeMemberLookupDTO{
			FullName:  user.FullName,
			AvatarURL: user.AvatarURL,
			Email:     user.Email,
			StudentID: user.StudentID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": members,
	})
}

func parseLoungeLookupFilters(r *http.Request) (service.AdminUserFilters, error) {
	query := r.URL.Query()

	var searchField, searchValue string
	for _, field := range []string{"full_name", "student_id", "email"} {
		values, present := query[field]
		if !present {
			continue
		}
		if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			return service.AdminUserFilters{}, fmt.Errorf("%s must be provided once with a non-empty value", field)
		}
		if searchField != "" {
			return service.AdminUserFilters{}, errors.New("provide exactly one of full_name, student_id, or email")
		}

		searchField = field
		searchValue = strings.TrimSpace(values[0])
	}

	if searchField == "" {
		return service.AdminUserFilters{}, errors.New("provide exactly one of full_name, student_id, or email")
	}

	filters := service.AdminUserFilters{Limit: 5}
	switch searchField {
	case "full_name":
		filters.FullName = searchValue
	case "student_id":
		filters.StudentID = searchValue
	case "email":
		filters.Email = searchValue
	}

	return filters, nil
}

func (h *LoungeHandler) writeCreateSessionError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetReqID(r.Context())
	switch {
	case errors.Is(err, service.ErrInvalidLoungeSession):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_LOUNGE_SESSION", "The lounge session request is invalid.", requestID)
	case errors.Is(err, service.ErrNoLoungeAccess):
		util.WriteApiResponse(w, http.StatusForbidden, "NO_LOUNGE_ACCESS", service.ErrNoLoungeAccess.Error(), requestID)
	case errors.Is(err, service.ErrActiveSessionExists):
		util.WriteApiResponse(w, http.StatusConflict, "ACTIVE_SESSION_EXISTS", service.ErrActiveSessionExists.Error(), requestID)
	case errors.Is(err, service.ErrStationClosed):
		util.WriteApiResponse(w, http.StatusConflict, "STATION_CLOSED", service.ErrStationClosed.Error(), requestID)
	case errors.Is(err, service.ErrStationInUse):
		util.WriteApiResponse(w, http.StatusConflict, "STATION_IN_USE", service.ErrStationInUse.Error(), requestID)
	case errors.Is(err, service.ErrLoungeStationNotFound):
		util.WriteApiResponse(w, http.StatusNotFound, "STATION_NOT_FOUND", service.ErrLoungeStationNotFound.Error(), requestID)
	default:
		slog.ErrorContext(r.Context(), "unable to create lounge session",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create lounge session.", requestID)
	}
}
