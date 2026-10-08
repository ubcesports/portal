package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type MembershipInvitationHandler struct {
	service *service.MembershipInvitationService
}

func NewMembershipInvitationHandler(
	service *service.MembershipInvitationService,
) *MembershipInvitationHandler {
	return &MembershipInvitationHandler{service: service}
}

/*
Creates a paid membership purchase for a person who does not yet have an
account, then sends an invitation to the supplied email address.

API URL: POST /admin/membership-invitations

Args (JSON body):

	email: email address that must be used when creating the account
	tier_id: UUID of the purchased membership tier
	amount_paid_cents: nonnegative amount received, in cents
	payment_method: offline payment method (cash or etransfer)

Returns:

	the created membership invitation (HTTP 201)

Raises:

	400: invalid request body, email, tier ID, amount, or payment method
	401: user is not authenticated
	403: user is not an admin
	409: an invitation or user account already exists for the email address
	500: the purchase could not be created or its invitation could not be prepared
*/
func (h *MembershipInvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	createdByUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var request dto.CreateMembershipInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body.", requestID)
		return
	}

	invitation, err := h.service.Create(r.Context(), request, createdByUserID, requestID)
	if err != nil {
		h.writeError(w, r, err, "create")
		return
	}
	util.WriteJson(w, http.StatusCreated, invitation)
}

/*
Returns every membership invitation, including completed invitations, newest
first.

API URL: GET /admin/membership-invitations

Args:

	None

Returns:

	an array of pending and completed membership invitations (HTTP 200)

Raises:

	401: user is not authenticated
	403: user is not an admin
	500: membership invitations could not be loaded
*/
func (h *MembershipInvitationHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	invitations, err := h.service.GetAll(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to list membership invitations",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load membership invitations.", requestID)
		return
	}
	util.WriteJson(w, http.StatusOK, invitations)
}

/*
Returns one unclaimed membership invitation.

API URL: GET /admin/membership-invitations/{id}

Args:

	id: membership invitation UUID, from the URL path

Returns:

	the requested membership invitation (HTTP 200)

Raises:

	400: malformed invitation ID
	401: user is not authenticated
	403: user is not an admin
	404: membership invitation does not exist
	500: membership invitation could not be loaded
*/
func (h *MembershipInvitationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	invitation, err := h.service.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, r, err, "load")
		return
	}
	util.WriteJson(w, http.StatusOK, invitation)
}

/*
Deletes an unclaimed membership invitation without creating a membership or
transaction.

API URL: DELETE /admin/membership-invitations/{id}

Args:

	id: membership invitation UUID, from the URL path

Returns:

	no response body (HTTP 204)

Raises:

	400: malformed invitation ID
	401: user is not authenticated
	403: user is not an admin
	404: membership invitation does not exist
	500: membership invitation could not be deleted
*/
func (h *MembershipInvitationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		h.writeError(w, r, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MembershipInvitationHandler) writeError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	action string,
) {
	requestID := middleware.GetReqID(r.Context())
	switch {
	case errors.Is(err, service.ErrInvalidMembershipInvitation):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_MEMBERSHIP_INVITATION", "The membership invitation is invalid.", requestID)
	case errors.Is(err, service.ErrInvalidMembershipInvitationTier):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_MEMBERSHIP_TIER", "The membership tier is invalid.", requestID)
	case errors.Is(err, service.ErrOfflinePaymentMethod):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_PAYMENT_METHOD", service.ErrOfflinePaymentMethod.Error(), requestID)
	case errors.Is(err, service.ErrMembershipInvitationUserExists):
		util.WriteApiResponse(w, http.StatusConflict, "USER_ALREADY_EXISTS", service.ErrMembershipInvitationUserExists.Error(), requestID)
	case errors.Is(err, service.ErrMembershipInvitationEmail):
		util.WriteApiResponse(w, http.StatusConflict, "MEMBERSHIP_INVITATION_EXISTS", service.ErrMembershipInvitationEmail.Error(), requestID)
	case errors.Is(err, service.ErrMembershipInvitationNotFound):
		util.WriteApiResponse(w, http.StatusNotFound, "MEMBERSHIP_INVITATION_NOT_FOUND", service.ErrMembershipInvitationNotFound.Error(), requestID)
	default:
		slog.ErrorContext(r.Context(), "unable to "+action+" membership invitation",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to "+action+" membership invitation.", requestID)
	}
}
