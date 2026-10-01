package handlers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/service"
	"github.com/ubcesports/memberships/internal/util"
)

type AdminHandler struct {
	adminService      *service.AdminService
	membershipService *service.MembershipService
}

/*
	Public functions
*/

func NewAdminHandler(adminService *service.AdminService, membershipService *service.MembershipService) *AdminHandler {
	return &AdminHandler{adminService: adminService, membershipService: membershipService}
}

/*
Returns a filtered and paginated list of users.

API URL: GET /admin/users

Args (query params):

	full_name: optional case-insensitive name substring
	student_id: optional case-insensitive student ID substring
	email: optional case-insensitive email substring
	role: optional role (member or admin)
	is_student: optional boolean student status
	group: optional repeated group values; matches the user's exact group set
	membership_tier_id: optional repeated tier UUIDs; matches the user's exact active tier set
	limit: optional page size (default 25, maximum 100)
	offset: optional number of users to skip (default 0)

Returns:

	users: paginated user profiles (HTTP 200)
	total: number of users matching the filters

Raises:

	400: invalid filter or pagination value
	401: user is not authenticated
	403: user is not an admin
	500: users could not be loaded
*/
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	filters, err := parseAdminUserFilters(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, total, err := h.adminService.GetUsers(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load users",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		http.Error(w, "unable to load users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"users": users,
		"total": total,
	})
}

/*
Returns membership tier options used by the admin user filter.

API URL: GET /admin/membership-tiers

Args:

	None

Returns:

	[]dto.AdminMembershipTierOption (HTTP 200)

Raises:

	401: user is not authenticated
	403: user is not an admin
	500: membership tier options could not be retrieved
*/
func (h *AdminHandler) GetMembershipTierOptions(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	options, err := h.adminService.GetAdminMembershipTierOptions(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load admin membership tier options",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load membership tier options.", requestID)
		return
	}

	util.WriteJson(w, http.StatusOK, options)
}

/*
Exports every user matching the supplied filters as CSV.

API URL: GET /admin/users/export

Args (query params):

	full_name: optional case-insensitive name substring
	student_id: optional case-insensitive student ID substring
	email: optional case-insensitive email substring
	role: optional role (member or admin)
	is_student: optional boolean student status
	group: optional repeated group values; matches the user's exact group set
	membership_tier_id: optional repeated tier UUIDs; matches the user's exact active tier set

Returns:

	users.csv: CSV file containing all matching users (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: users could not be exported
*/
func (h *AdminHandler) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	filters, err := parseAdminUserFilters(r, false)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "BAD_REQUEST", "Filters to get users could not be parsed.", requestId)
		return
	}

	users, err := h.adminService.ExportUsers(r.Context(), filters, userId, requestId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to export users",
			"error", err,
			"request_id", requestId,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to export users.", requestId)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="users.csv"`)

	writer := csv.NewWriter(w)
	if err := writer.Write([]string{
		"ID",
		"Full Name",
		"Email",
		"Student ID",
		"Role",
		"Is Student",
		"Groups",
		"Active Memberships",
		"Created At",
		"Updated At",
		"Email Verified At",
		"Onboarding Completed At",
		"Avatar URL",
	}); err != nil {
		slog.ErrorContext(r.Context(), "unable to write CSV header",
			"error", err,
			"request_id", requestId,
		)
		return
	}

	for _, user := range users {
		groups := make([]string, 0, len(user.Groups))
		for _, group := range user.Groups {
			groups = append(groups, string(group))
		}
		membershipTitles := make([]string, 0, len(user.ActiveMemberships))
		for _, membership := range user.ActiveMemberships {
			membershipTitles = append(membershipTitles, membership.TierTitle)
		}

		if err := writer.Write([]string{
			user.ID,
			safeCSVCell(user.FullName),
			safeCSVCell(user.Email),
			safeCSVCell(optionalString(user.StudentID)),
			string(user.Role),
			strconv.FormatBool(user.IsStudent),
			strings.Join(groups, ";"),
			strings.Join(membershipTitles, ";"),
			user.CreatedAt.Format(time.RFC3339),
			user.UpdatedAt.Format(time.RFC3339),
			optionalTime(user.EmailVerifiedAt),
			optionalTime(user.OnboardingCompletedAt),
			safeCSVCell(optionalString(user.AvatarURL)),
		}); err != nil {
			slog.ErrorContext(r.Context(), "unable to write CSV row",
				"error", err,
				"request_id", requestId,
			)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		slog.ErrorContext(r.Context(), "unable to flush CSV response",
			"error", err,
			"request_id", requestId,
		)
		return
	}
}

/*
Returns a single user's profile.

API URL: GET /admin/users/{id}

Args:

	id: the user's UUID, from the URL path

Returns:

	response body containing the profile under the "user" key (HTTP 200)

Raises:

	400: malformed user ID
	401: user is not authenticated
	403: user is not an admin
	404: user does not exist
	500: the user could not be loaded
*/
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	profile, err := h.adminService.GetUserByID(r.Context(), targetUserId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestId)
			return
		}

		slog.ErrorContext(r.Context(), "unable to load user",
			"error", err,
			"request_id", requestId,
			"user_id", targetUserId,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load user.", requestId)
		return
	}

	util.WriteJson(w, http.StatusOK, map[string]dto.ProfileDTO{"user": *profile})
}

/*
Returns every membership the user has held, newest first, each with the
transaction that paid for it.

API URL: GET /admin/users/{id}/memberships

Args:

	id: the user's UUID, from the URL path

Returns:

	response body containing an array of memberships (HTTP 200). A user who has
	never completed a checkout has none, which is an empty array, not a 404.

Raises:

	400: malformed user ID
	401: user is not authenticated
	403: user is not an admin
	404: user does not exist
	500: the memberships could not be loaded
*/
func (h *AdminHandler) GetUserMemberships(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	memberships, err := h.adminService.GetUserMemberships(r.Context(), targetUserId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestId)
			return
		}

		slog.ErrorContext(r.Context(), "unable to load user memberships",
			"error", err,
			"request_id", requestId,
			"user_id", targetUserId,
		)
		util.WriteApiResponse(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Unable to load memberships.",
			requestId,
		)
		return
	}

	util.WriteJson(w, http.StatusOK, memberships)
}

/*
Updates the editable fields on a single user.

API URL: PATCH /admin/users/{id}

Args (JSON body, every field optional):

	full_name: new nonblank full name
	student_id: new student ID, only editable while the user is a student
	is_student: new student status
	groups_add: groups to add to the user
	groups_remove: groups to remove from the user (member is always kept)
	role: new role (member or admin), empty string floors the user to member
	cancel_membership_id: nullable membership id. not null means that membership will be cancelled

Returns:

	response body containing the updated profile under the "user" key (HTTP 200)

Raises:

	400: invalid request body or validation error
	401: user is not authenticated
	403: user is not an admin
	404: target user does not exist
	409: conflict, e.g. reinstating while another membership is active
	500: the user could not be updated
*/
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	actorId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	var updateUserRequest dto.AdminUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	profile, err := h.adminService.UpdateUser(
		r.Context(),
		actorId,
		targetUserId,
		requestId,
		buildUpdateUserRequest(updateUserRequest),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			util.WriteApiResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestId)

		case errors.Is(err, service.ErrNotFound):
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestId)

		case errors.Is(err, service.ErrConflict):
			util.WriteApiResponse(w, http.StatusConflict, "CONFLICT", err.Error(), requestId)

		default:
			slog.ErrorContext(r.Context(), "unable to update user",
				"error", err,
				"request_id", requestId,
				"user_id", targetUserId,
			)
			util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to update user. Please try again.", requestId)
		}

		return
	}

	util.WriteJson(w, http.StatusOK, map[string]dto.ProfileDTO{"user": *profile})
}

/*
Updates the editable fields of the executive profile of a single user.

API URL: PATCH /admin/exec-profile/{id}

Args (JSON body, every field optional):

	title: new title for the executive profile
	display_order: new display order for the executive profile
	display_group: new display group for the executive profile

Returns:

	response body containing the updated executive profile under the "exec_profile" key (HTTP 200)

Raises:

	400: invalid request body or validation error
	401: user is not authenticated
	403: user is not an admin
	404: target user does not exist
	500: the executive profile could not be updated
*/
func (h *AdminHandler) UpdateExecProfile(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	actorId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	var updateExecProfileRequest dto.AdminUpdateExecProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&updateExecProfileRequest); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	profile, err := h.adminService.UpdateExecProfile(
		r.Context(),
		actorId,
		targetUserId,
		util.ToPgText(updateExecProfileRequest.Title),
		util.ToPgInt4(updateExecProfileRequest.DisplayOrder),
		util.ToNullGroupType(updateExecProfileRequest.DisplayGroup),
		requestId,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			util.WriteApiResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), requestId)

		case errors.Is(err, service.ErrNotFound):
			util.WriteApiResponse(w, http.StatusNotFound, "NOT_FOUND", err.Error(), requestId)

		default:
			slog.ErrorContext(r.Context(), "unable to update exec profile",
				"error", err,
				"request_id", requestId,
				"user_id", targetUserId,
			)
			util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to update exec profile. Please try again.", requestId)
		}

		return
	}

	util.WriteJson(w, http.StatusOK, map[string]db.GetExecProfileByUserIDRow{"exec_profile": profile})
}

/*
Returns a paginated list of admin audit logs.

API URL: GET /admin/audit-logs

Args (query params):

	actor_name: optional case-insensitive actor name substring
	limit: optional page size (default 25, maximum 100)
	offset: optional number of logs to skip (default 0)

Returns:

	logs: paginated admin audit logs (HTTP 200)

Raises:

	400: invalid pagination value
	401: user is not authenticated
	403: user is not an admin
	500: audit logs could not be loaded for some reason
*/
func (h *AdminHandler) GetAdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	filters, err := parseAdminAuditLogFilters(r)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), requestID)
		return
	}

	logs, total, err := h.adminService.GetAdminAuditLogs(r.Context(), filters)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load admin audit logs",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Unable to load admin audit logs. Please try again.",
			requestID,
		)
		return
	}

	util.WriteJson(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

/*
Exports every audit log matching the supplied filters as CSV.

API URL: GET /admin/audit-logs/export

Args (query params):

	actor_name: optional case-insensitive actor name substring

Returns:

	audit-logs.csv: CSV file containing all matching audit logs (HTTP 200)

Raises:

	400: invalid filter value
	401: user is not authenticated
	403: user is not an admin
	500: audit logs could not be exported
*/
func (h *AdminHandler) ExportAuditLogsCSV(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	userId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	filters, err := parseAdminAuditLogFilters(r)
	if err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "BAD_REQUEST", "Filters to get users could not be parsed.", requestId)
		return
	}

	logs, err := h.adminService.ExportAuditLogs(r.Context(), filters, userId, requestId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to export users",
			"error", err,
			"request_id", requestId,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to export users.", requestId)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="users.csv"`)

	writer := csv.NewWriter(w)
	if err := writer.Write([]string{
		"Actor User ID",
		"Actor",
		"Avatar URL",
		"Occurred At",
		"Action",
		"Description",
		"Outcome",
		"Request ID",
		"Target User",
		"Target User Avatar URL",
	}); err != nil {
		slog.ErrorContext(r.Context(), "unable to write CSV header",
			"error", err,
			"request_id", requestId,
		)
		return
	}

	for _, log := range logs {
		description := log.Description
		if description == nil {
			description = new(string)
		}
		if err := writer.Write([]string{
			log.Actor.ActorUserId,
			log.Actor.ActorFullName,
			safeCSVCell(log.Actor.ActorAvatarURL),
			log.OccuredAt.Format(time.RFC3339),
			log.Action,
			safeCSVCell(*description),
			string(log.Outcome),
			log.RequestId,
			safeCSVCell(log.TargetUser.ActorFullName),
			safeCSVCell(log.TargetUser.ActorAvatarURL),
		}); err != nil {
			slog.ErrorContext(r.Context(), "unable to write CSV row",
				"error", err,
				"request_id", requestId,
			)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		slog.ErrorContext(r.Context(), "unable to flush CSV response",
			"error", err,
			"request_id", requestId,
		)
		return
	}
}

/*
Creates a new cash/etransfer membership for a user

API URL: POST /admin/membership/add/{id}

Args:

	id (query param): user id of person to add the membership to
	AdminAddMembershipToUserRequest (request body): as seen in dto/admin.go

Returns:

	None: (HTTP 200)

Raises:

	400: invalid request body, user ID, tier ID, or payment method
	401: user is not authenticated
	403: user is not an admin
	403: selected membership tier is unavailable or purchases are closed
	409: an already-paid checkout is still being processed
	500: membership could not be added to user
*/
func (h *AdminHandler) AddMembershipToUser(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	// Get current user id
	actorId, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	var addMembershipRequest dto.AdminAddMembershipToUserRequest
	if err := json.NewDecoder(r.Body).Decode(&addMembershipRequest); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body. Please try again.", requestId)
		return
	}

	err := h.membershipService.AddMembershipToUser(
		r.Context(),
		actorId,
		targetUserId,
		requestId,
		addMembershipRequest,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOfflinePaymentMethod):
			util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_PAYMENT_METHOD", service.ErrOfflinePaymentMethod.Error(), requestId)

		case errors.Is(err, service.ErrInvalidMembershipTier):
			util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_TIER_ID", service.ErrInvalidMembershipTier.Error(), requestId)

		case errors.Is(err, service.ErrTierNotEligible):
			util.WriteApiResponse(w, http.StatusForbidden, "TIER_NOT_AVAILABLE", service.ErrTierNotEligible.Error(), requestId)

		case errors.Is(err, service.ErrMembershipPurchaseClosed):
			util.WriteApiResponse(w, http.StatusForbidden, "MEMBERSHIP_PURCHASE_CLOSED", service.ErrMembershipPurchaseClosed.Error(), requestId)

		case errors.Is(err, service.ErrPendingCheckoutAlreadyPaid):
			util.WriteApiResponse(w, http.StatusConflict, "CHECKOUT_ALREADY_PAID", service.ErrPendingCheckoutAlreadyPaid.Error(), requestId)

		default:
			slog.ErrorContext(r.Context(), "unable to add membership to user",
				"error", err,
				"request_id", requestId,
				"user_id", targetUserId,
			)
			util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to add membership. Please try again.", requestId)
		}

		return
	}

	util.WriteJson(w, http.StatusOK, nil)
}

/*
Returns membership tiers and prices the target user

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

API URL: GET /admin/membership/eligible/{id}

Args:

	auth.Session user id
	id (query param): id of the user to fetch the eligible memberships of

Returns:

	[]dto.EligibleMembershipTierDTO (HTTP 200)

Raises:

	401: user is not authenticated
	500: eligible membership tiers and prices could not be retrieved
*/
func (h *AdminHandler) GetEligibleTiersWithPricesById(w http.ResponseWriter, r *http.Request) {
	requestId := middleware.GetReqID(r.Context())

	_, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestId)
		return
	}

	targetUserId := chi.URLParam(r, "id")
	if _, err := util.GetValidatedUUID(targetUserId); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID.", requestId)
		return
	}

	tiers, err := h.membershipService.GetEligibleTiersWithPrices(r.Context(), targetUserId)
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to load eligible membership tiers", "error", err, "request_id", requestId, "user_id", targetUserId)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load eligible membership tiers", requestId)
		return
	}

	if tiers != nil {
		util.WriteJson(w, 200, tiers)
	} else {
		util.WriteJson(w, 200, nil)
	}
}

/*
	Private functions
*/

func buildUpdateUserRequest(request dto.AdminUpdateUserRequest) service.UpdateUserRequest {
	updateUserRequest := service.UpdateUserRequest{
		FullName:           request.FullName,
		StudentID:          request.StudentID,
		IsStudent:          request.IsStudent,
		GroupsAdd:          make([]db.GroupType, 0, len(request.GroupsAdd)),
		GroupsRemove:       make([]db.GroupType, 0, len(request.GroupsRemove)),
		CancelMembershipId: request.CancelMembershipId,
	}

	if request.Role != nil {
		role := db.RoleType(*request.Role)
		updateUserRequest.Role = &role
	}

	for _, group := range request.GroupsAdd {
		updateUserRequest.GroupsAdd = append(updateUserRequest.GroupsAdd, db.GroupType(group))
	}
	for _, group := range request.GroupsRemove {
		updateUserRequest.GroupsRemove = append(updateUserRequest.GroupsRemove, db.GroupType(group))
	}

	return updateUserRequest
}

func parseAdminAuditLogFilters(r *http.Request) (service.AdminAuditLogFilters, error) {
	query := r.URL.Query()
	filters := service.AdminAuditLogFilters{
		ActorName: query.Get("actor_name"),
		Limit:     25,
	}

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return service.AdminAuditLogFilters{}, errors.New("limit must be a positive integer")
		}
		filters.Limit = int32(parsed)
	}

	if value := query.Get("offset"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 0 {
			return service.AdminAuditLogFilters{}, errors.New("offset must be a non-negative integer")
		}
		filters.Offset = int32(parsed)
	}

	return filters, nil
}

func parseAdminUserFilters(r *http.Request, includePagination bool) (service.AdminUserFilters, error) {
	query := r.URL.Query()
	filters := service.AdminUserFilters{
		FullName:          query.Get("full_name"),
		StudentID:         query.Get("student_id"),
		Email:             query.Get("email"),
		Role:              query.Get("role"),
		Groups:            query["group"],
		MembershipTierIDs: query["membership_tier_id"],
		Limit:             25,
	}

	if value := query.Get("is_student"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return service.AdminUserFilters{}, errors.New("is_student must be true or false")
		}
		filters.IsStudent = &parsed
	}

	if filters.Role != "" {
		switch dto.RoleType(filters.Role) {
		case dto.RoleMember, dto.RoleAdmin:
		default:
			return service.AdminUserFilters{}, errors.New("invalid role")
		}
	}

	for _, group := range filters.Groups {
		switch dto.GroupType(group) {
		case dto.GroupMember,
			dto.GroupCompetitiveTeam,
			dto.GroupExecutive,
			dto.GroupDirector,
			dto.GroupBoard,
			dto.GroupPresident:
		default:
			return service.AdminUserFilters{}, errors.New("invalid group")
		}
	}

	for _, tierID := range filters.MembershipTierIDs {
		if _, err := util.GetValidatedUUID(tierID); err != nil {
			return service.AdminUserFilters{}, errors.New("invalid membership tier ID")
		}
	}

	if !includePagination {
		filters.Limit = 0
		return filters, nil
	}

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return service.AdminUserFilters{}, errors.New("limit must be a positive integer")
		}
		filters.Limit = int32(parsed)
	}

	if value := query.Get("offset"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 0 {
			return service.AdminUserFilters{}, errors.New("offset must be a non-negative integer")
		}
		filters.Offset = int32(parsed)
	}

	return filters, nil
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func safeCSVCell(value string) string {
	if value == "" || !strings.ContainsRune("=+-@", rune(value[0])) {
		return value
	}
	return "'" + value
}
