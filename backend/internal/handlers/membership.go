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

type MembershipHandler struct {
	membershipService *service.MembershipService
}

func NewMembershipHandler(membershipService *service.MembershipService) *MembershipHandler {
	return &MembershipHandler{membershipService: membershipService}
}

/*
Returns all public membership tiers and prices.

-> Day pass with student/non-student prices
-> Basic pass with student/non-student prices
-> Lounge pass with student/non-student prices

API URL: GET /membership/tiers

Args:

	None

Returns:

	[]dto.MembershipTierDTO (HTTP 200)

Raises:

	500: public membership tiers and prices could not be retrieved
*/
func (h *MembershipHandler) GetPublicTiersWithPrices(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	tiers, err := h.membershipService.GetPublicTiersAndPrices(r.Context())
	util.SetPublicCacheHeader(w, 60, 300, 60)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load public membership tiers", "error", err, "request_id", requestID)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load membership tiers", requestID)
		return
	}
	util.WriteJson(w, 200, tiers)
}

/*
Returns membership tiers and prices the current user is eligible to buy.

For exec/director/board members:

	Executive Pass

For competitive team players:

	Competitive Team Pass

For regular UBC students:

	Day pass with student price
	Basic pass with student price
	Lounge pass with student price

For regular non-students:

	Day pass with non-student price
	Basic pass with non-student price
	Lounge pass with non-student price

API URL: GET /membership/tiers/eligible

Args:

	auth.Session user id

Returns:

	[]dto.EligibleMembershipTierDTO (HTTP 200)

Raises:

	401: user is not authenticated
	500: eligible membership tiers and prices could not be retrieved
*/
func (h *MembershipHandler) GetEligibleTiersWithPrices(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	tiers, err := h.membershipService.GetEligibleTiersWithPrices(r.Context(), userId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load eligible membership tiers", "error", err, "request_id", requestID, "user_id", userId)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load eligible membership tiers", requestID)
		return
	}

	if tiers != nil {
		util.WriteJson(w, 200, tiers)
	} else {
		util.WriteJson(w, 200, nil)
	}
}

/*
Returns the current user's active memberships and transactions.

API URL: GET /membership/me/current

Args:

	auth.Session user id

Returns:

	[]dto.MembershipDTO (HTTP 200)

Raises:

	401: user is not authenticated
	500: current membership could not be retrieved
*/
func (h *MembershipHandler) GetCurrentMembershipWithTransaction(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	membership, err := h.membershipService.GetCurrentMembershipsWithTransactions(r.Context(), userId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load current memberships", "error", err, "request_id", requestID, "user_id", userId)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load current memberships", requestID)
		return
	}
	util.WriteJson(w, 200, membership)
}

/*
Returns all past and present memberships and transactions for the current user.

API URL: GET /membership/me/all

Args:

	auth.Session user id

Returns:

	[]dto.MembershipDTO (HTTP 200)

Raises:

	401: user is not authenticated
	500: memberships could not be retrieved
*/
func (h *MembershipHandler) GetAllMembershipsWithTransactions(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	memberships, err := h.membershipService.GetAllMembershipsWithTransactions(r.Context(), userId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load memberships", "error", err, "request_id", requestID, "user_id", userId)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load memberships", requestID)
		return
	}
	util.WriteJson(w, 200, memberships)
}

/*
Creates a signed Stripe Checkout Session for a membership purchase.

API URL: POST /membership/checkout

Args:

	auth.Session user id
	dto.CheckoutSessionRequest

Returns:

	dto.CheckoutSessionResponse (HTTP 200)

Raises:

	400: request body is unreadable or malformed
	401: user is not authenticated
	403: requested tier is not eligible for the user
	404: requested tier does not exist
	409: user already has a membership that blocks this purchase
	500: checkout session could not be created
*/
func (h *MembershipHandler) CreateMembershipCheckoutSession(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var checkoutSessionRequest dto.CheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&checkoutSessionRequest); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestID)
		return
	}

	checkoutSessionResponse, err := h.membershipService.CreateCheckoutSession(r.Context(), userId, checkoutSessionRequest)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMembershipAlreadyExists):
			util.WriteApiResponse(w, http.StatusConflict, "CONFLICT", err.Error(), requestID)
			return

		case errors.Is(err, service.ErrTierNotEligible):
			util.WriteApiResponse(w, http.StatusForbidden, "TIER_NOT_AVAILABLE", err.Error(), requestID)
			return

		case errors.Is(err, service.ErrTierNotFound):
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestID)
			return

		case errors.Is(err, service.ErrMembershipPurchaseClosed):
			util.WriteApiResponse(w, http.StatusForbidden, "MEMBERSHIP_PURCHASE_CLOSED", err.Error(), requestID)
			return

		case errors.Is(err, service.ErrPendingCheckoutAlreadyPaid):
			util.WriteApiResponse(w, http.StatusConflict, "CHECKOUT_ALREADY_PAID", err.Error(), requestID)
			return

		default:
			slog.ErrorContext(r.Context(), "unable to create membership checkout session", "error", err, "request_id", requestID, "user_id", userId)
			util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create checkout session", requestID)
		}

		return
	}

	util.WriteJson(w, http.StatusOK, *checkoutSessionResponse)
}
