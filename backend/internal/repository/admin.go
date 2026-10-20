package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/util"
)

// AdminStore is the set of persistence operations the admin service depends on.
//
// It exists so the service can be exercised without a database. The interface
// lives here rather than in the service package because WithTx has to hand the
// callback a store of this same type.
type AdminStore interface {
	UpdateUserFullName(ctx context.Context, userId string, fullName string) error
	GetUsers(ctx context.Context, params db.GetUsersAdminParams) ([]db.GetUsersAdminRow, error)
	CountUsers(ctx context.Context, params db.CountUsersAdminParams) (int64, error)
	GetAdminMembershipTierOptions(ctx context.Context) ([]db.GetAdminMembershipTierOptionsRow, error)
	CountAdminAuditLogs(ctx context.Context, params pgtype.Text) (int64, error)
	CreateAdminAuditLog(ctx context.Context, params db.CreateAdminAuditLogParams) error
	GetAdminAuditLogs(ctx context.Context, params db.GetAdminAuditLogsParams) ([]db.GetAdminAuditLogsRow, error)
	GetUserByID(ctx context.Context, userId string) (db.GetAdminUserByIDRow, error)
	UpdateUserStudentInfo(ctx context.Context, userId string, isStudent bool, studentId string) error
	StudentIDExists(ctx context.Context, studentId string) (bool, error)
	UpdateUserRole(ctx context.Context, userId string, role db.RoleType) error
	AddUserGroup(ctx context.Context, userId string, group db.GroupType) error
	CreateExecProfile(ctx context.Context, userId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) error
	HasExecProfile(ctx context.Context, userId string) (bool, error)
	HasExecGroup(ctx context.Context, userId string) (bool, error)
	RemoveExecProfile(ctx context.Context, userId string) error
	UpdateExecProfile(ctx context.Context, userId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) (db.GetExecProfileByUserIDRow, error)
	RemoveUserGroup(ctx context.Context, userId string, group db.GroupType) error
	GetUserMemberships(ctx context.Context, userId string) ([]db.GetAllMembershipsWithTransactionsRow, error)
	HasActiveMembership(ctx context.Context, userId string) (bool, error)
	CancelActiveMembershipByUserIdAndMembershipId(ctx context.Context, userId string, membershipId string, occurredAt time.Time) (bool, error)
	WithTx(ctx context.Context, fn func(AdminStore) error) error
}

func (r *AdminRepository) GetAdminMembershipTierOptions(
	ctx context.Context,
) ([]db.GetAdminMembershipTierOptionsRow, error) {
	rows, err := r.store.GetAdminMembershipTierOptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("query admin membership tier options: %w", err)
	}
	return rows, nil
}

type AdminRepository struct {
	pool  *pgxpool.Pool
	store *db.Queries
}

var _ AdminStore = (*AdminRepository)(nil)

func NewAdminRepository(pool *pgxpool.Pool, store *db.Queries) *AdminRepository {
	return &AdminRepository{pool: pool, store: store}
}

func (r *AdminRepository) GetUsers(
	ctx context.Context,
	params db.GetUsersAdminParams) ([]db.GetUsersAdminRow, error) {
	rows, err := r.store.GetUsersAdmin(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query admin users: %w", err)
	}
	return rows, nil
}

func (r *AdminRepository) CountUsers(
	ctx context.Context,
	params db.CountUsersAdminParams,
) (int64, error) {
	count, err := r.store.CountUsersAdmin(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count admin users: %w", err)
	}
	return count, nil
}

func (r *AdminRepository) CountAdminAuditLogs(
	ctx context.Context,
	params pgtype.Text,
) (int64, error) {
	count, err := r.store.CountAdminAuditLogs(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count admin audit logs: %w", err)
	}
	return count, nil
}

func (r *AdminRepository) CreateAdminAuditLog(ctx context.Context, params db.CreateAdminAuditLogParams) error {
	err := r.store.CreateAdminAuditLog(ctx, params)
	if err != nil {
		return fmt.Errorf("create admin audit log: %w", err)
	}
	return nil
}

func (r *AdminRepository) GetAdminAuditLogs(ctx context.Context, params db.GetAdminAuditLogsParams) ([]db.GetAdminAuditLogsRow, error) {
	logs, err := r.store.GetAdminAuditLogs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query admin audit logs: %w", err)
	}
	return logs, nil
}

func (r *AdminRepository) GetUserByID(ctx context.Context, userId string) (db.GetAdminUserByIDRow, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return db.GetAdminUserByIDRow{}, err
	}

	row, err := r.store.GetAdminUserByID(ctx, pgUserId)
	if err != nil {
		return db.GetAdminUserByIDRow{}, fmt.Errorf("query admin user by ID: %w", err)
	}
	return row, nil
}

func (r *AdminRepository) UpdateUserFullName(ctx context.Context, userId string, fullName string) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}
	if err := r.store.UpdateUserFullName(ctx, db.UpdateUserFullNameParams{ID: pgUserId, FullName: fullName}); err != nil {
		return fmt.Errorf("update user full name: %w", err)
	}
	return nil
}

func (r *AdminRepository) UpdateUserStudentInfo(
	ctx context.Context,
	userId string,
	isStudent bool,
	studentId string,
) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	err = r.store.UpdateUserStudentInfo(ctx, db.UpdateUserStudentInfoParams{
		ID:        pgUserId,
		IsStudent: isStudent,
		StudentID: pgtype.Text{
			String: studentId,
			Valid:  studentId != "",
		},
	})
	if err != nil {
		return fmt.Errorf("update user student info: %w", err)
	}
	return nil
}

func (r *AdminRepository) StudentIDExists(ctx context.Context, studentId string) (bool, error) {
	exists, err := r.store.StudentIDExists(ctx, pgtype.Text{
		String: studentId,
		Valid:  studentId != "",
	})
	if err != nil {
		return false, fmt.Errorf("check student ID availability: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) UpdateUserRole(ctx context.Context, userId string, role db.RoleType) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	err = r.store.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   pgUserId,
		Role: role,
	})
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}

func (r *AdminRepository) AddUserGroup(ctx context.Context, userId string, group db.GroupType) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	err = r.store.AddUserGroup(ctx, db.AddUserGroupParams{
		UserID: pgUserId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("add user group: %w", err)
	}
	return nil
}

func (r *AdminRepository) CreateExecProfile(ctx context.Context, userId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	err = r.store.CreateExecProfile(ctx, db.CreateExecProfileParams{
		UserID:       pgUserId,
		Title:        title,
		DisplayOrder: displayOrder,
		DisplayGroup: displayGroup,
	})
	if err != nil {
		return fmt.Errorf("create exec profile: %w", err)
	}
	return nil
}

func (r *AdminRepository) HasExecProfile(ctx context.Context, userId string) (bool, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return false, err
	}

	exists, err := r.store.HasExecProfileForUser(ctx, pgUserId)
	if err != nil {
		return false, fmt.Errorf("check existing exec profile: %w", err)
	}
	return exists, err
}

func (r *AdminRepository) HasExecGroup(ctx context.Context, userId string) (bool, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return false, err
	}

	exists, err := r.store.HasExecGroupForUser(ctx, pgUserId)
	if err != nil {
		return false, fmt.Errorf("check existing exec groups: %w", err)
	}
	return exists, err
}

func (r *AdminRepository) RemoveExecProfile(ctx context.Context, userId string) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}
	return r.store.RemoveExecProfile(ctx, pgUserId)
}

func (r *AdminRepository) UpdateExecProfile(ctx context.Context, userId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullExecDisplayGroupType) (db.GetExecProfileByUserIDRow, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, err
	}

	err = r.store.UpdateExecProfile(ctx, db.UpdateExecProfileParams{
		UserID:       pgUserId,
		Title:        title,
		DisplayOrder: displayOrder,
		DisplayGroup: displayGroup,
	})
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, fmt.Errorf("update exec profile: %w", err)
	}

	updatedProfile, err := r.store.GetExecProfileByUserID(ctx, pgUserId)
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, fmt.Errorf("get updated exec profile: %w", err)
	}
	return updatedProfile, nil
}

func (r *AdminRepository) RemoveUserGroup(ctx context.Context, userId string, group db.GroupType) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	err = r.store.RemoveUserGroup(ctx, db.RemoveUserGroupParams{
		UserID: pgUserId,
		Group:  group,
	})
	if err != nil {
		return fmt.Errorf("remove user group: %w", err)
	}
	return nil
}

// GetUserMemberships returns every membership the user has ever held, newest
// first, joined to the transaction that paid for it.
func (r *AdminRepository) GetUserMemberships(
	ctx context.Context,
	userId string,
) ([]db.GetAllMembershipsWithTransactionsRow, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return nil, err
	}

	rows, err := r.store.GetAllMembershipsWithTransactions(ctx, pgUserId)
	if err != nil {
		return nil, fmt.Errorf("query user memberships: %w", err)
	}
	return rows, nil
}

func (r *AdminRepository) HasActiveMembership(ctx context.Context, userId string) (bool, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return false, err
	}

	exists, err := r.store.HasActiveMembershipForUser(ctx, pgUserId)
	if err != nil {
		return false, fmt.Errorf("check active membership: %w", err)
	}
	return exists, nil
}

func (r *AdminRepository) CancelActiveMembershipByUserIdAndMembershipId(
	ctx context.Context,
	userId string,
	membershipId string,
	occurredAt time.Time,
) (bool, error) {
	pgUserID, err := util.GetValidatedUUID(userId)
	if err != nil {
		return false, err
	}

	pgMembershipID, err := util.GetValidatedUUID(membershipId)
	if err != nil {
		return false, err
	}

	rowsAffected, err :=
		r.store.CancelActiveMembershipByUserIdAndMembershipId(
			ctx,
			db.CancelActiveMembershipByUserIdAndMembershipIdParams{
				UserID: pgUserID,
				ID:     pgMembershipID,
				CancelledAt: pgtype.Timestamptz{
					Time:  occurredAt,
					Valid: true,
				},
			},
		)
	if err != nil {
		return false, fmt.Errorf(
			"cancel active membership: %w",
			err,
		)
	}

	return rowsAffected == 1, nil
}

// executes fn within a database transaction.
//
// The callback receives an AdminStore whose operations are executed using the
// same transaction. If fn returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// This helper should be used when multiple repository operations must succeed
// or fail atomically.
func (r *AdminRepository) WithTx(ctx context.Context, fn func(AdminStore) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &AdminRepository{
		pool:  r.pool,
		store: r.store.WithTx(tx),
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
