package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/util"
)

// RoleStore is the set of persistence operations the role service depends on,
// so the service can be tested without a database.
type RoleStore interface {
	GetAll(ctx context.Context) ([]dto.RoleResponse, error)
	GetByID(ctx context.Context, id string) (*dto.RoleDetailsResponse, error)
	GetForUser(ctx context.Context, userID string) ([]dto.UserRolePermissions, error)
	Create(ctx context.Context, request dto.CreateRoleRequest, createdByUserID string) (string, error)
	Update(ctx context.Context, id string, request dto.UpdateRoleRequest) error
	Delete(ctx context.Context, id string) error
	GetAssigneeName(ctx context.Context, userID string) (string, error)
	AssignUser(ctx context.Context, roleID string, userID string, assignedByUserID string) (bool, error)
	UnassignUser(ctx context.Context, roleID string, userID string) (bool, error)
	CreateAuditLog(ctx context.Context, params RoleAuditLogParams) error
	WithTx(ctx context.Context, fn func(RoleStore) error) error
}

type RoleRepository struct {
	pool  *pgxpool.Pool
	store *db.Queries
}

type RoleAuditLogParams struct {
	ActorUserID string
	Action      string
	Outcome     db.AdminAuditOutcomeType
	RequestID   string
	Description string
}

var _ RoleStore = (*RoleRepository)(nil)

func NewRoleRepository(pool *pgxpool.Pool, store *db.Queries) *RoleRepository {
	return &RoleRepository{pool: pool, store: store}
}

func (r *RoleRepository) GetAll(ctx context.Context) ([]dto.RoleResponse, error) {
	rows, err := r.store.GetRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("get roles: %w", err)
	}

	roles := make([]dto.RoleResponse, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, roleResponse(
			row.ID,
			row.Name,
			row.Description,
			row.Color,
			row.DisplayOrder,
			row.CreatedByUserID,
			row.CreatedByName,
			row.CreatedByProfilePicture,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}

	return roles, nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id string) (*dto.RoleDetailsResponse, error) {
	roleID, err := util.GetValidatedUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	return r.getByID(ctx, roleID)
}

// Create inserts a role and its permission keys, returning the new role's ID.
// Run it inside WithTx.
func (r *RoleRepository) Create(ctx context.Context, request dto.CreateRoleRequest, createdByUserID string) (string, error) {
	creatorID, err := util.GetValidatedUUID(createdByUserID)
	if err != nil {
		return "", fmt.Errorf("invalid creator user ID: %w", err)
	}

	displayOrder := int32(0)
	if request.DisplayOrder != nil {
		displayOrder = *request.DisplayOrder
	}

	roleID, err := r.store.CreateRole(ctx, db.CreateRoleParams{
		Name:            request.Name,
		Description:     util.ToPgText(request.Description),
		Color:           util.ToPgText(request.Color),
		DisplayOrder:    displayOrder,
		CreatedByUserID: creatorID,
	})
	if err != nil {
		return "", fmt.Errorf("create role: %w", err)
	}

	if err := r.addPermissionKeys(ctx, roleID, request.PermissionKeys); err != nil {
		return "", err
	}

	return roleID.String(), nil
}

// Update applies the non-nil fields of request. A set PermissionKeys replaces
// the role's permissions. Run it inside WithTx.
func (r *RoleRepository) Update(ctx context.Context, id string, request dto.UpdateRoleRequest) error {
	roleID, err := util.GetValidatedUUID(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	if _, err := r.store.UpdateRole(ctx, db.UpdateRoleParams{
		Name:         util.ToPgText(request.Name),
		Description:  util.ToPgText(request.Description),
		Color:        util.ToPgText(request.Color),
		DisplayOrder: util.ToPgInt4(request.DisplayOrder),
		ID:           roleID,
	}); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	if request.PermissionKeys != nil {
		if err := r.store.DeleteRolePermissionKeys(ctx, roleID); err != nil {
			return fmt.Errorf("delete role permission keys: %w", err)
		}
		if err := r.addPermissionKeys(ctx, roleID, *request.PermissionKeys); err != nil {
			return err
		}
	}

	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	roleID, err := util.GetValidatedUUID(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	rowsAffected, err := r.store.DeleteRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *RoleRepository) GetForUser(ctx context.Context, userID string) ([]dto.UserRolePermissions, error) {
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	rows, err := r.store.GetRolesForUser(ctx, pgUserID)
	if err != nil {
		return nil, fmt.Errorf("get roles for user: %w", err)
	}

	roles := make([]dto.UserRolePermissions, 0)
	roleIndexes := make(map[string]int)
	for _, row := range rows {
		roleID := row.ID.String()
		index, exists := roleIndexes[roleID]
		if !exists {
			index = len(roles)
			roleIndexes[roleID] = index
			roles = append(roles, dto.UserRolePermissions{
				ID:             roleID,
				Name:           row.Name,
				PermissionKeys: make([]string, 0),
			})
		}

		if row.Permission.Valid {
			roles[index].PermissionKeys = append(roles[index].PermissionKeys, row.Permission.String)
		}
	}

	return roles, nil
}

// GetAssigneeName returns the user's full name, or pgx.ErrNoRows.
func (r *RoleRepository) GetAssigneeName(ctx context.Context, userID string) (string, error) {
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	row, err := r.store.GetRoleAssignee(ctx, pgUserID)
	if err != nil {
		return "", err
	}
	return row.FullName, nil
}

// AssignUser reports false when the user already had the role.
func (r *RoleRepository) AssignUser(ctx context.Context, roleID string, userID string, assignedByUserID string) (bool, error) {
	pgRoleID, err := util.GetValidatedUUID(roleID)
	if err != nil {
		return false, fmt.Errorf("invalid role ID: %w", err)
	}
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}
	pgAssignerID, err := util.GetValidatedUUID(assignedByUserID)
	if err != nil {
		return false, fmt.Errorf("invalid assigner user ID: %w", err)
	}

	rowsAffected, err := r.store.AssignUserRole(ctx, db.AssignUserRoleParams{
		UserID:           pgUserID,
		RoleID:           pgRoleID,
		AssignedByUserID: pgAssignerID,
	})
	if err != nil {
		return false, fmt.Errorf("assign user role: %w", err)
	}
	return rowsAffected > 0, nil
}

// UnassignUser reports false when the user did not have the role.
func (r *RoleRepository) UnassignUser(ctx context.Context, roleID string, userID string) (bool, error) {
	pgRoleID, err := util.GetValidatedUUID(roleID)
	if err != nil {
		return false, fmt.Errorf("invalid role ID: %w", err)
	}
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	rowsAffected, err := r.store.UnassignUserRole(ctx, db.UnassignUserRoleParams{
		UserID: pgUserID,
		RoleID: pgRoleID,
	})
	if err != nil {
		return false, fmt.Errorf("unassign user role: %w", err)
	}
	return rowsAffected > 0, nil
}

// HasAnyRole satisfies auth.RoleChecker.
func (r *RoleRepository) HasAnyRole(ctx context.Context, userID string, roleNames []string) (bool, error) {
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	hasRole, err := r.store.UserHasAnyRole(ctx, db.UserHasAnyRoleParams{
		UserID:    pgUserID,
		RoleNames: roleNames,
	})
	if err != nil {
		return false, fmt.Errorf("check user roles: %w", err)
	}
	return hasRole, nil
}

// GetPermissionsForUser satisfies auth.PermissionLoader.
func (r *RoleRepository) GetPermissionsForUser(ctx context.Context, userID string) ([]string, error) {
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	permissions, err := r.store.GetPermissionsForUser(ctx, pgUserID)
	if err != nil {
		return nil, fmt.Errorf("get permissions for user: %w", err)
	}
	return permissions, nil
}

/*
	Helpers
*/

func (r *RoleRepository) getByID(ctx context.Context, roleID pgtype.UUID) (*dto.RoleDetailsResponse, error) {
	row, err := r.store.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role by ID: %w", err)
	}

	usersJSON, err := json.Marshal(row.UsersWithRole)
	if err != nil {
		return nil, fmt.Errorf("encode users with role: %w", err)
	}
	users := make([]dto.RoleUser, 0)
	if err := json.Unmarshal(usersJSON, &users); err != nil {
		return nil, fmt.Errorf("decode users with role: %w", err)
	}

	return &dto.RoleDetailsResponse{
		RoleResponse: roleResponse(
			row.ID,
			row.Name,
			row.Description,
			row.Color,
			row.DisplayOrder,
			row.CreatedByUserID,
			row.CreatedByName,
			row.CreatedByProfilePicture,
			row.CreatedAt,
			row.UpdatedAt,
		),
		PermissionKeys: row.PermissionKeys,
		UsersWithRole:  users,
	}, nil
}

func (r *RoleRepository) addPermissionKeys(ctx context.Context, roleID pgtype.UUID, permissionKeys []string) error {
	if len(permissionKeys) == 0 {
		return nil
	}

	if _, err := r.store.AddRolePermissionKeys(ctx, db.AddRolePermissionKeysParams{
		RoleID:         roleID,
		PermissionKeys: permissionKeys,
	}); err != nil {
		return fmt.Errorf("add role permission keys: %w", err)
	}
	return nil
}

func (r *RoleRepository) CreateAuditLog(ctx context.Context, params RoleAuditLogParams) error {
	actorUserID, err := util.GetValidatedUUID(params.ActorUserID)
	if err != nil {
		return fmt.Errorf("invalid audit actor user ID: %w", err)
	}

	return r.store.CreateAdminAuditLog(ctx, db.CreateAdminAuditLogParams{
		ActorUserID: actorUserID,
		Action:      params.Action,
		Outcome:     params.Outcome,
		RequestID:   params.RequestID,
		Description: pgtype.Text{
			String: params.Description,
			Valid:  params.Description != "",
		},
	})
}

// WithTx runs role operations, and their audit entries, in one transaction.
func (r *RoleRepository) WithTx(ctx context.Context, fn func(RoleStore) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &RoleRepository{pool: r.pool, store: r.store.WithTx(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func roleResponse(
	id pgtype.UUID,
	name string,
	description pgtype.Text,
	color pgtype.Text,
	displayOrder int32,
	createdByUserID pgtype.UUID,
	createdByName pgtype.Text,
	createdByProfilePicture pgtype.Text,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) dto.RoleResponse {
	var createdBy *dto.RoleUser
	if createdByUserID.Valid {
		createdBy = &dto.RoleUser{
			UserID:         createdByUserID.String(),
			Name:           createdByName.String,
			ProfilePicture: util.TextPointer(createdByProfilePicture),
		}
	}

	return dto.RoleResponse{
		ID:           id.String(),
		Name:         name,
		Description:  util.TextPointer(description),
		Color:        util.TextPointer(color),
		DisplayOrder: displayOrder,
		CreatedBy:    createdBy,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
}
