package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type LoungeHandler struct {
	loungeService *service.LoungeService
}

func NewLoungeHandler(loungeService *service.LoungeService) *LoungeHandler {
	return &LoungeHandler{loungeService: loungeService}
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
Checks out an active lounge session.

API URL: POST /lounge/sessions/{session_id}/checkout

Args:

	session_id: UUID of the session in URL path
	JSON body (optional):
		checked_out_by: optional UUID of the user performing the checkout

Returns:

	the updated lounge session (HTTP 200)

Raises:

	400: invalid session ID or request body
	404: session not found (SESSION_NOT_FOUND)
	409: session already checked out (SESSION_ALREADY_CHECKED_OUT)
	500: unable to checkout session (INTERNAL_ERROR)
*/
func (h *LoungeHandler) CheckoutSession(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	sessionID := chi.URLParam(r, "session_id")
	if _, err := util.GetValidatedUUID(sessionID); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid session ID.", requestID)
		return
	}

	var checkedOutBy *string
	if r.Body != nil && r.ContentLength != 0 {
		var req dto.CheckoutSessionRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body.", requestID)
			return
		}
		if req.CheckedOutBy != nil {
			if _, err := util.GetValidatedUUID(*req.CheckedOutBy); err != nil {
				util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_ID", "Invalid checked_out_by ID.", requestID)
				return
			}
			checkedOutBy = req.CheckedOutBy
		}
	}

	if checkedOutBy == nil {
		if actorID, ok := util.CurrentUserID(r); ok {
			checkedOutBy = &actorID
		}
	}

	updated, err := h.loungeService.CheckoutSession(r.Context(), sessionID, checkedOutBy)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			util.WriteApiResponse(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found.", requestID)
			return
		}
		if errors.Is(err, service.ErrSessionAlreadyCheckedOut) {
			util.WriteApiResponse(w, http.StatusConflict, "SESSION_ALREADY_CHECKED_OUT", "Session already checked out.", requestID)
			return
		}
		slog.ErrorContext(r.Context(), "unable to checkout session",
			"error", err,
			"session_id", sessionID,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to checkout session.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, updated)
}
