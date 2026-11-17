package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

var (
	ErrRoleNotFound          = errors.New("Role not found")
	ErrRoleNameTaken         = errors.New("A role with this name already exists")
	ErrInvalidRole           = errors.New("Invalid role")
	ErrInvalidRolePermission = errors.New("Invalid role permission")
	ErrRoleUserNotFound      = errors.New("User not found")
	ErrRoleAlreadyAssigned   = errors.New("This user already has this role")
	ErrRoleNotAssigned       = errors.New("This user does not have this role")
)

const (
	actionRoleCreated    = "role.created"
	actionRoleEdited     = "role.updated"
	actionRoleDeleted    = "role.deleted"
	actionRoleAssigned   = "role.assigned"
	actionRoleUnassigned = "role.unassigned"

	maxRoleNameLength = 50
)

var roleColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type RoleService struct {
	repository repository.RoleStore
}

func NewRoleService(repository *repository.RoleRepository) *RoleService {
	return &RoleService{repository: repository}
}

func (s *RoleService) GetAll(ctx context.Context) ([]dto.RoleResponse, error) {
	return s.repository.GetAll(ctx)
}

func (s *RoleService) GetByID(ctx context.Context, id string) (*dto.RoleDetailsResponse, error) {
	if _, err := util.GetValidatedUUID(id); err != nil {
		return nil, ErrInvalidRole
	}
	role, err := s.repository.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRoleNotFound
	}
	return role, err
}

// GetForUser returns the user's roles with their permission keys.
func (s *RoleService) GetForUser(ctx context.Context, userID string) ([]dto.UserRolePermissions, error) {
	return s.repository.GetForUser(ctx, userID)
}

func (s *RoleService) Create(
	ctx context.Context,
	request dto.CreateRoleRequest,
	actorUserID string,
	requestID string,
) (*dto.RoleDetailsResponse, error) {
	failureDescription := fmt.Sprintf("Failed to create role %s", strings.TrimSpace(request.Name))

	normalized, err := normalizeCreateRole(request)
	if err != nil {
		return nil, s.auditRoleFailure(ctx, actorUserID, actionRoleCreated, requestID, failureDescription, err)
	}

	var role *dto.RoleDetailsResponse
	err = s.repository.WithTx(ctx, func(repo repository.RoleStore) error {
		roleID, err := repo.Create(ctx, normalized, actorUserID)
		if err != nil {
			return err
		}

		role, err = repo.GetByID(ctx, roleID)
		if err != nil {
			return err
		}

		return repo.CreateAuditLog(ctx, repository.RoleAuditLogParams{
			ActorUserID: actorUserID,
			Action:      actionRoleCreated,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: fmt.Sprintf(
				"Created role %s with permissions: %s",
				role.Name,
				formatPermissionKeys(role.PermissionKeys),
			),
		})
	})
	if err != nil {
		return nil, s.auditRoleFailure(ctx, actorUserID, actionRoleCreated, requestID, failureDescription, roleError(err))
	}

	return role, nil
}

func (s *RoleService) Update(
	ctx context.Context,
	id string,
	request dto.UpdateRoleRequest,
	actorUserID string,
	requestID string,
) (*dto.RoleDetailsResponse, error) {
	failureDescription := fmt.Sprintf("Failed to update role %s", id)

	if _, err := util.GetValidatedUUID(id); err != nil {
		return nil, s.auditRoleFailure(ctx, actorUserID, actionRoleEdited, requestID, failureDescription, ErrInvalidRole)
	}
	normalized, err := normalizeUpdateRole(request)
	if err != nil {
		return nil, s.auditRoleFailure(ctx, actorUserID, actionRoleEdited, requestID, failureDescription, err)
	}

	var role *dto.RoleDetailsResponse
	err = s.repository.WithTx(ctx, func(repo repository.RoleStore) error {
		existing, err := repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to update role %s", existing.Name)

		if err := repo.Update(ctx, id, normalized); err != nil {
			return err
		}

		role, err = repo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		return repo.CreateAuditLog(ctx, repository.RoleAuditLogParams{
			ActorUserID: actorUserID,
			Action:      actionRoleEdited,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: describeRoleUpdate(existing, role),
		})
	})
	if err != nil {
		return nil, s.auditRoleFailure(ctx, actorUserID, actionRoleEdited, requestID, failureDescription, roleError(err))
	}

	return role, nil
}

func (s *RoleService) Delete(
	ctx context.Context,
	id string,
	actorUserID string,
	requestID string,
) error {
	failureDescription := fmt.Sprintf("Failed to delete role %s", id)

	if _, err := util.GetValidatedUUID(id); err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleDeleted, requestID, failureDescription, ErrInvalidRole)
	}

	err := s.repository.WithTx(ctx, func(repo repository.RoleStore) error {
		existing, err := repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to delete role %s", existing.Name)

		if err := repo.Delete(ctx, id); err != nil {
			return err
		}

		return repo.CreateAuditLog(ctx, repository.RoleAuditLogParams{
			ActorUserID: actorUserID,
			Action:      actionRoleDeleted,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: fmt.Sprintf(
				"Deleted role %s (removed from %d users)",
				existing.Name,
				len(existing.UsersWithRole),
			),
		})
	})
	if err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleDeleted, requestID, failureDescription, roleError(err))
	}

	return nil
}

// AssignUser gives a user a role they do not already have.
func (s *RoleService) AssignUser(
	ctx context.Context,
	roleID string,
	request dto.AssignRoleRequest,
	actorUserID string,
	requestID string,
) error {
	request.UserID = strings.TrimSpace(request.UserID)
	failureDescription := fmt.Sprintf("Failed to assign role %s to user %s", roleID, request.UserID)

	if _, err := util.GetValidatedUUID(roleID); err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleAssigned, requestID, failureDescription, ErrInvalidRole)
	}
	if _, err := util.GetValidatedUUID(request.UserID); err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleAssigned, requestID, failureDescription, ErrInvalidRole)
	}

	err := s.repository.WithTx(ctx, func(repo repository.RoleStore) error {
		role, err := repo.GetByID(ctx, roleID)
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to assign role %s to user %s", role.Name, request.UserID)

		userName, err := repo.GetAssigneeName(ctx, request.UserID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoleUserNotFound
		}
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to assign role %s to %s", role.Name, userName)

		assigned, err := repo.AssignUser(ctx, roleID, request.UserID, actorUserID)
		if err != nil {
			return err
		}
		if !assigned {
			return ErrRoleAlreadyAssigned
		}

		return repo.CreateAuditLog(ctx, repository.RoleAuditLogParams{
			ActorUserID: actorUserID,
			Action:      actionRoleAssigned,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: fmt.Sprintf("Assigned role %s to %s", role.Name, userName),
		})
	})
	if err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleAssigned, requestID, failureDescription, roleError(err))
	}
	return nil
}

// UnassignUser removes a role the user currently has.
func (s *RoleService) UnassignUser(
	ctx context.Context,
	roleID string,
	userID string,
	actorUserID string,
	requestID string,
) error {
	userID = strings.TrimSpace(userID)
	failureDescription := fmt.Sprintf("Failed to remove role %s from user %s", roleID, userID)

	if _, err := util.GetValidatedUUID(roleID); err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleUnassigned, requestID, failureDescription, ErrInvalidRole)
	}
	if _, err := util.GetValidatedUUID(userID); err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleUnassigned, requestID, failureDescription, ErrInvalidRole)
	}

	err := s.repository.WithTx(ctx, func(repo repository.RoleStore) error {
		role, err := repo.GetByID(ctx, roleID)
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to remove role %s from user %s", role.Name, userID)

		userName, err := repo.GetAssigneeName(ctx, userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoleUserNotFound
		}
		if err != nil {
			return err
		}
		failureDescription = fmt.Sprintf("Failed to remove role %s from %s", role.Name, userName)

		removed, err := repo.UnassignUser(ctx, roleID, userID)
		if err != nil {
			return err
		}
		if !removed {
			return ErrRoleNotAssigned
		}

		return repo.CreateAuditLog(ctx, repository.RoleAuditLogParams{
			ActorUserID: actorUserID,
			Action:      actionRoleUnassigned,
			Outcome:     db.AdminAuditOutcomeTypeSuccess,
			RequestID:   requestID,
			Description: fmt.Sprintf("Removed role %s from %s", role.Name, userName),
		})
	})
	if err != nil {
		return s.auditRoleFailure(ctx, actorUserID, actionRoleUnassigned, requestID, failureDescription, roleError(err))
	}
	return nil
}

/*
	Private functions
*/

func normalizeCreateRole(request dto.CreateRoleRequest) (dto.CreateRoleRequest, error) {
	name, err := normalizeRoleName(request.Name)
	if err != nil {
		return request, err
	}
	request.Name = name

	request.Description = normalizeRoleDescription(request.Description)

	if request.Color, err = normalizeRoleColor(request.Color); err != nil {
		return request, err
	}
	if err := validateRoleDisplayOrder(request.DisplayOrder); err != nil {
		return request, err
	}
	if request.PermissionKeys, err = normalizePermissionKeys(request.PermissionKeys); err != nil {
		return request, err
	}

	return request, nil
}

func normalizeUpdateRole(request dto.UpdateRoleRequest) (dto.UpdateRoleRequest, error) {
	if request.Name == nil &&
		request.Description == nil &&
		request.Color == nil &&
		request.DisplayOrder == nil &&
		request.PermissionKeys == nil {
		return request, ErrInvalidRole
	}

	if request.Name != nil {
		name, err := normalizeRoleName(*request.Name)
		if err != nil {
			return request, err
		}
		request.Name = &name
	}

	if request.Description != nil {
		description := strings.TrimSpace(*request.Description)
		request.Description = &description
	}

	var err error
	if request.Color, err = normalizeRoleColor(request.Color); err != nil {
		return request, err
	}
	if err := validateRoleDisplayOrder(request.DisplayOrder); err != nil {
		return request, err
	}
	if request.PermissionKeys != nil {
		keys, err := normalizePermissionKeys(*request.PermissionKeys)
		if err != nil {
			return request, err
		}
		request.PermissionKeys = &keys
	}

	return request, nil
}

func normalizeRoleName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxRoleNameLength {
		return "", ErrInvalidRole
	}
	return name, nil
}

// normalizeRoleDescription trims the description; blank becomes nil.
func normalizeRoleDescription(description *string) *string {
	if description == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*description)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeRoleColor(color *string) (*string, error) {
	if color == nil {
		return nil, nil
	}
	normalized := strings.ToUpper(strings.TrimSpace(*color))
	if !roleColorPattern.MatchString(normalized) {
		return nil, ErrInvalidRole
	}
	return &normalized, nil
}

func validateRoleDisplayOrder(displayOrder *int32) error {
	if displayOrder != nil && *displayOrder < 0 {
		return ErrInvalidRole
	}
	return nil
}

// normalizePermissionKeys trims and de-duplicates keys, keeping order. Unknown
// keys fail the role_permissions foreign key (see roleError).
func normalizePermissionKeys(keys []string) ([]string, error) {
	normalized := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, ErrInvalidRolePermission
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	return normalized, nil
}

func roleError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrRoleNameTaken
		case "23503":
			return ErrInvalidRolePermission
		case "23514", "23502", "22001":
			return ErrInvalidRole
		}
	}
	return err
}

func describeRoleUpdate(before *dto.RoleDetailsResponse, after *dto.RoleDetailsResponse) string {
	changes := make([]string, 0)
	if before.Name != after.Name {
		changes = append(changes, fmt.Sprintf("name %q → %q", before.Name, after.Name))
	}
	if stringValue(before.Description) != stringValue(after.Description) {
		changes = append(changes, "description")
	}
	if stringValue(before.Color) != stringValue(after.Color) {
		changes = append(changes, fmt.Sprintf("color %s → %s", formatOptional(before.Color), formatOptional(after.Color)))
	}
	if before.DisplayOrder != after.DisplayOrder {
		changes = append(changes, fmt.Sprintf("display order %d → %d", before.DisplayOrder, after.DisplayOrder))
	}
	if formatPermissionKeys(before.PermissionKeys) != formatPermissionKeys(after.PermissionKeys) {
		changes = append(changes, fmt.Sprintf(
			"permissions %s → %s",
			formatPermissionKeys(before.PermissionKeys),
			formatPermissionKeys(after.PermissionKeys),
		))
	}

	if len(changes) == 0 {
		return fmt.Sprintf("Updated role %s (no changes)", after.Name)
	}
	return fmt.Sprintf("Updated role %s: %s", after.Name, strings.Join(changes, "; "))
}

// formatPermissionKeys expects sorted keys, as GetRoleByID returns them.
func formatPermissionKeys(keys []string) string {
	if len(keys) == 0 {
		return "none"
	}
	return strings.Join(keys, ", ")
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatOptional(value *string) string {
	if value == nil {
		return "none"
	}
	return *value
}

func (s *RoleService) auditRoleFailure(
	ctx context.Context,
	actorUserID string,
	action string,
	requestID string,
	description string,
	operationErr error,
) error {
	auditErr := s.repository.CreateAuditLog(ctx, repository.RoleAuditLogParams{
		ActorUserID: actorUserID,
		Action:      action,
		Outcome:     db.AdminAuditOutcomeTypeFailed,
		RequestID:   requestID,
		Description: description,
	})
	if auditErr != nil {
		return errors.Join(operationErr, auditErr)
	}
	return operationErr
}
