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

type RoleHandler struct {
	service *service.RoleService
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

/*
Returns every role ordered by display order, then name.

API URL: GET /admin/roles

Args:

	None

Returns:

	an array of roles, without their permissions or assigned users (HTTP 200)

Raises:

	401: user is not authenticated
	403: user is not an admin
	500: roles could not be loaded
*/
func (h *RoleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	roles, err := h.service.GetAll(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "unable to list roles",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load roles.", requestID)
		return
	}
	util.WriteJson(w, http.StatusOK, roles)
}

/*
Returns one role with its permission keys and the users assigned to it.

API URL: GET /admin/roles/{id}

Args:

	id: role UUID, from the URL path

Returns:

	the requested role (HTTP 200)

Raises:

	400: malformed role ID
	401: user is not authenticated
	403: user is not an admin
	404: role does not exist
	500: role could not be loaded
*/
func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	role, err := h.service.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, r, err, "load")
		return
	}
	util.WriteJson(w, http.StatusOK, role)
}

/*
Creates a role with the given permissions.

API URL: POST /admin/roles

Args (JSON body):

	name: unique role name, 1-50 characters
	description: optional description
	color: optional hex color in the form #RRGGBB
	display_order: optional nonnegative sort order (defaults to 0)
	permission_keys: permission keys granted by the role, eg. "users.view"

Returns:

	the created role with its permissions and assigned users (HTTP 201)

Raises:

	400: invalid request body, role fields, or permission keys
	401: user is not authenticated
	403: user is not an admin
	409: a role with this name already exists
	500: role could not be created
*/
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	actorUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var request dto.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body.", requestID)
		return
	}

	role, err := h.service.Create(r.Context(), request, actorUserID, requestID)
	if err != nil {
		h.writeError(w, r, err, "create")
		return
	}
	util.WriteJson(w, http.StatusCreated, role)
}

/*
Updates a role. Only the fields present in the body are changed. When
permission_keys is present, it replaces the role's permissions; an empty array
removes them all.

API URL: PATCH /admin/roles/{id}

Args:

	id: role UUID, from the URL path

Args (JSON body, at least one required):

	name: unique role name, 1-50 characters
	description: description
	color: hex color in the form #RRGGBB
	display_order: nonnegative sort order
	permission_keys: the complete set of permission keys for the role

Returns:

	the updated role with its permissions and assigned users (HTTP 200)

Raises:

	400: malformed role ID, invalid request body, role fields, or permission keys
	401: user is not authenticated
	403: user is not an admin
	404: role does not exist
	409: a role with this name already exists
	500: role could not be updated
*/
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	actorUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var request dto.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body.", requestID)
		return
	}

	role, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), request, actorUserID, requestID)
	if err != nil {
		h.writeError(w, r, err, "update")
		return
	}
	util.WriteJson(w, http.StatusOK, role)
}

/*
Deletes a role, removing it from every user it was assigned to.

API URL: DELETE /admin/roles/{id}

Args:

	id: role UUID, from the URL path

Returns:

	no response body (HTTP 204)

Raises:

	400: malformed role ID
	401: user is not authenticated
	403: user is not an admin
	404: role does not exist
	500: role could not be deleted
*/
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	actorUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id"), actorUserID, requestID); err != nil {
		h.writeError(w, r, err, "delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/*
Assigns a role to a user.

API URL: POST /admin/roles/{id}/users

Args:

	id: role UUID, from the URL path

Args (JSON body):

	user_id: UUID of the user to give the role to

Returns:

	no response body (HTTP 204)

Raises:

	400: malformed role or user ID, or invalid request body
	401: user is not authenticated
	403: user is not an admin
	404: role or user does not exist
	409: the user already has the role
	500: role could not be assigned
*/
func (h *RoleHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	actorUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	var request dto.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body.", requestID)
		return
	}

	if err := h.service.AssignUser(r.Context(), chi.URLParam(r, "id"), request, actorUserID, requestID); err != nil {
		h.writeError(w, r, err, "assign")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/*
Removes a role from a user.

API URL: DELETE /admin/roles/{id}/users/{userId}

Args:

	id: role UUID, from the URL path
	userId: user UUID, from the URL path

Returns:

	no response body (HTTP 204)

Raises:

	400: malformed role or user ID
	401: user is not authenticated
	403: user is not an admin
	404: role or user does not exist, or the user does not have the role
	500: role could not be removed
*/
func (h *RoleHandler) UnassignUser(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	actorUserID, ok := util.CurrentUserID(r)
	if !ok {
		util.WriteApiResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}

	if err := h.service.UnassignUser(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "userId"), actorUserID, requestID); err != nil {
		h.writeError(w, r, err, "unassign")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) writeError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	action string,
) {
	requestID := middleware.GetReqID(r.Context())
	switch {
	case errors.Is(err, service.ErrInvalidRole):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_ROLE", "The role is invalid.", requestID)
	case errors.Is(err, service.ErrInvalidRolePermission):
		util.WriteApiResponse(w, http.StatusBadRequest, "INVALID_ROLE_PERMISSION", "One or more permissions are invalid.", requestID)
	case errors.Is(err, service.ErrRoleNameTaken):
		util.WriteApiResponse(w, http.StatusConflict, "ROLE_NAME_TAKEN", service.ErrRoleNameTaken.Error(), requestID)
	case errors.Is(err, service.ErrRoleUserNotFound):
		util.WriteApiResponse(w, http.StatusNotFound, "USER_NOT_FOUND", service.ErrRoleUserNotFound.Error(), requestID)
	case errors.Is(err, service.ErrRoleNotAssigned):
		util.WriteApiResponse(w, http.StatusNotFound, "ROLE_NOT_ASSIGNED", service.ErrRoleNotAssigned.Error(), requestID)
	case errors.Is(err, service.ErrRoleAlreadyAssigned):
		util.WriteApiResponse(w, http.StatusConflict, "ROLE_ALREADY_ASSIGNED", service.ErrRoleAlreadyAssigned.Error(), requestID)
	case errors.Is(err, service.ErrRoleNotFound):
		util.WriteApiResponse(w, http.StatusNotFound, "ROLE_NOT_FOUND", service.ErrRoleNotFound.Error(), requestID)
	default:
		slog.ErrorContext(r.Context(), "unable to "+action+" role",
			"error", err,
			"request_id", requestID,
		)
		util.WriteApiResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to "+action+" role.", requestID)
	}
}
