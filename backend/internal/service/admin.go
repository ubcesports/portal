package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/mailer"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

// Postgres error code for a unique constraint violation.
const pgUniqueViolationCode = "23505"

var defaultDisplayGroupType = map[db.GroupType]db.ExecDisplayGroupType{
	"executive": db.ExecDisplayGroupTypeExecutive,
	"board":     db.ExecDisplayGroupTypeBoard,
	"director":  db.ExecDisplayGroupTypeGameDirector,
	"president": db.ExecDisplayGroupTypePresident,
}

type AdminUserFilters struct {
	FullName          string
	StudentID         string
	Email             string
	Role              string
	IsStudent         *bool
	Groups            []string
	MembershipTierIDs []string
	Limit             int32
	Offset            int32
}

type AdminAuditLogFilters struct {
	ActorName string
	Limit     int32
	Offset    int32
}

type AdminAuditLogInput struct {
	ActorUserID  string
	Action       string
	TargetUserID string
	Outcome      db.AdminAuditOutcomeType
	RequestID    string
	Description  string
}

// UpdateUserRequest describes the edits an admin wants to apply to a user.
// Every field is optional; only the ones that are set are acted on.
type UpdateUserRequest struct {
	FullName           *string
	StudentID          *string
	IsStudent          *bool
	GroupsAdd          []db.GroupType
	GroupsRemove       []db.GroupType
	Role               *db.RoleType
	CancelMembershipId *string
}

// Audit log actions emitted by UpdateUser.
const (
	actionUserUpdated          = "user.updated"
	actionFullNameUpdated      = "user.full_name.updated"
	actionStudentIDUpdated     = "user.student_id.updated"
	actionStudentStatusUpdated = "user.student_status.updated"
	actionRoleUpdated          = "user.role.updated"
	actionGroupAdded           = "user.group.added"
	actionGroupRemoved         = "user.group.removed"
	actionMembershipCancelled  = "user.membership.cancelled"
)

// Number of times a random non-student ID is regenerated before giving up.
const maxNonStudentIDAttempts = 5

// pendingAuditLog is an audit entry that has been earned by a successful action
// but not yet written.
type pendingAuditLog struct {
	action      string
	description string
	email       *pendingUserEmail // nil = no email for this entry
}

// pendingUserEmail is a fully rendered-content email earned by a successful
// admin action, sent only after the enclosing transaction has committed.
type pendingUserEmail struct {
	heading    string
	subheading string
	rows       []mailer.Row
}

// auditableError attributes a failure to the action that caused it, so a
// "failed" audit entry can name it after the transaction has rolled back.
type auditableError struct {
	action      string
	description string
	err         error
}

func (e *auditableError) Error() string { return e.err.Error() }

func (e *auditableError) Unwrap() error { return e.err }

func auditable(action string, description string, err error) error {
	return &auditableError{action: action, description: description, err: err}
}

type AdminService struct {
	adminRepository repository.AdminStore
}

/*
	Public functions
*/

func NewAdminService(adminRepository *repository.AdminRepository) *AdminService {
	return &AdminService{adminRepository: adminRepository}
}

func (s *AdminService) GetUsers(ctx context.Context, filters AdminUserFilters) ([]dto.AdminUserDTO, int64, error) {
	if filters.Limit <= 0 {
		filters.Limit = 25
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	params := buildAdminQueryParams(filters)
	params.Limit = pgtype.Int4{Int32: filters.Limit, Valid: true}
	params.Offset = pgtype.Int4{Int32: filters.Offset, Valid: true}

	total, err := s.adminRepository.CountUsers(ctx, db.CountUsersAdminParams{
		FullName:          params.FullName,
		StudentID:         params.StudentID,
		Email:             params.Email,
		Role:              params.Role,
		IsStudent:         params.IsStudent,
		Groups:            params.Groups,
		MembershipTierIds: params.MembershipTierIds,
	})
	if err != nil {
		return nil, 0, err
	}

	users, err := s.getUsers(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *AdminService) ExportUsers(
	ctx context.Context,
	filters AdminUserFilters,
	actorId string,
	requestId string,
) ([]dto.AdminUserDTO, error) {
	users, exportErr := s.getUsers(ctx, buildAdminQueryParams(filters))

	outcome := db.AdminAuditOutcomeTypeSuccess
	description := fmt.Sprintf("Exported %d users", len(users))

	if exportErr != nil {
		outcome = db.AdminAuditOutcomeTypeFailed
		description = "Failed to export users"
	}

	auditErr := s.createAdminAuditLog(ctx, s.adminRepository, AdminAuditLogInput{
		ActorUserID: actorId,
		Action:      "users.exported",
		Outcome:     outcome,
		RequestID:   requestId,
		Description: description,
	})

	if auditErr != nil {
		if exportErr != nil {
			return nil, errors.Join(exportErr, auditErr)
		}

		return nil, auditErr
	}

	if exportErr != nil {
		return nil, exportErr
	}

	return users, nil
}

func (s *AdminService) GetAdminMembershipTierOptions(
	ctx context.Context,
) ([]dto.AdminMembershipTierOption, error) {
	rows, err := s.adminRepository.GetAdminMembershipTierOptions(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]dto.AdminMembershipTierOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, dto.AdminMembershipTierOption{
			ID:          row.ID.String(),
			Title:       row.Title,
			ProgramName: row.ProgramName,
		})
	}
	return options, nil
}

// GetUserByID returns a single user's profile for the admin detail view.
func (s *AdminService) GetUserByID(ctx context.Context, userId string) (*dto.ProfileDTO, error) {
	row, err := s.adminRepository.GetUserByID(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", ErrNotFound)
		}
		return nil, err
	}

	profile := buildAdminUserProfile(row)
	return &profile, nil
}

func (s *AdminService) UpdateExecProfile(ctx context.Context, actorId string, targetId string, title pgtype.Text, displayOrder pgtype.Int4, displayGroup db.NullGroupType, requestId string) (db.GetExecProfileByUserIDRow, error) {
	nullDisplayGroup, err := toNullExecDisplayGroupType(displayGroup)
	if err != nil {
		err = fmt.Errorf("%w: invalid display group", ErrValidation)
		outcome := db.AdminAuditOutcomeTypeFailed
		description := fmt.Sprintf("Failed to update exec profile for user %s", targetId)

		auditErr := s.createAdminAuditLog(ctx, s.adminRepository, AdminAuditLogInput{
			ActorUserID:  actorId,
			Action:       "exec_profile.updated",
			TargetUserID: targetId,
			Outcome:      outcome,
			RequestID:    requestId,
			Description:  description,
		})

		if auditErr != nil {
			if err != nil {
				return db.GetExecProfileByUserIDRow{}, errors.Join(err, auditErr)
			}
			return db.GetExecProfileByUserIDRow{}, auditErr
		}

		return db.GetExecProfileByUserIDRow{}, err
	}

	updatedProfile, err := s.adminRepository.UpdateExecProfile(ctx, targetId, title, displayOrder, nullDisplayGroup)
	if errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("%w: exec profile not found", ErrNotFound)
	}

	description := fmt.Sprintf("Updated exec profile for user %s", targetId)
	outcome := db.AdminAuditOutcomeTypeSuccess

	if err != nil {
		description = fmt.Sprintf("Failed to update exec profile for user %s", targetId)
		outcome = db.AdminAuditOutcomeTypeFailed
	}

	auditErr := s.createAdminAuditLog(ctx, s.adminRepository, AdminAuditLogInput{
		ActorUserID:  actorId,
		Action:       "exec_profile.updated",
		TargetUserID: targetId,
		Outcome:      outcome,
		RequestID:    requestId,
		Description:  description,
	})

	if auditErr != nil {
		if err != nil {
			return db.GetExecProfileByUserIDRow{}, errors.Join(err, auditErr)
		}
		return db.GetExecProfileByUserIDRow{}, auditErr
	}

	if err != nil {
		return db.GetExecProfileByUserIDRow{}, err
	}

	return updatedProfile, nil
}

// GetUserMemberships returns every membership the user has held, newest first.
//
// A user with no memberships is a normal state — they exist from sign up but
// only gain a membership by completing a checkout — so this returns an empty
// slice rather than an error.
func (s *AdminService) GetUserMemberships(ctx context.Context, userId string) ([]dto.MembershipDTO, error) {
	// Distinguishes "user does not exist" from "user has never bought one".
	if _, err := s.adminRepository.GetUserByID(ctx, userId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: user not found", ErrNotFound)
		}
		return nil, err
	}

	rows, err := s.adminRepository.GetUserMemberships(ctx, userId)
	if err != nil {
		return nil, err
	}

	memberships := make([]dto.MembershipDTO, 0, len(rows))
	for _, row := range rows {
		memberships = append(memberships, dto.MembershipDTO{
			ID:          row.ID.String(),
			TierId:      row.TierID.String(),
			TierTitle:   row.TierTitle,
			StartedAt:   row.StartedAt.Time,
			ExpiresAt:   row.ExpiresAt.Time,
			CancelledAt: util.TimestampPointer(row.CancelledAt),
			ProgramName: row.ProgramName,
			ProgramId:   row.ProgramID.String(),
			Transaction: dto.TransactionDTO{
				ID:                    row.TransactionID.String(),
				AmountPaid:            fmt.Sprintf("%.2f", float64(row.AmountPaidCents.Int64)/100),
				Status:                dto.TransactionStatusType(row.Status),
				GroupAtPurchase:       dto.GroupType(row.GroupAtPurchase.GroupType),
				StudentAtPurchase:     row.StudentAtPurchase.Bool,
				StripePaymentIntentId: row.StripePaymentIntentID.String,
				PurchaseType:          dto.PurchaseType(row.PurchaseType.PurchaseType),
				PaymentMethod:         dto.PaymentMethodType(row.PaymentMethod),
			},
		})
	}

	return memberships, nil
}

// UpdateUser applies every edit in req to the target user inside a single
// transaction and returns the user's resulting profile.
//
// Each individual change produces its own audit log entry. If any change fails
// the whole update is rolled back and a single "failed" entry naming the
// offending action is written instead.
func (s *AdminService) UpdateUser(
	ctx context.Context,
	actorId string,
	targetUserId string,
	requestId string,
	req UpdateUserRequest,
) (*dto.ProfileDTO, error) {
	var profile *dto.ProfileDTO
	var pendingEmails []pendingUserEmail
	var targetEmail string

	updateErr := s.adminRepository.WithTx(ctx, func(store repository.AdminStore) error {
		user, err := store.GetUserByID(ctx, targetUserId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return auditable(
					actionUserUpdated,
					"Failed to update user: user not found",
					fmt.Errorf("%w: user not found", ErrNotFound),
				)
			}
			return err
		}
		targetEmail = user.Email

		entries, err := s.applyUserUpdates(ctx, store, user, req)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if err := s.createAdminAuditLog(ctx, store, AdminAuditLogInput{
				ActorUserID:  actorId,
				Action:       entry.action,
				TargetUserID: targetUserId,
				Outcome:      db.AdminAuditOutcomeTypeSuccess,
				RequestID:    requestId,
				Description:  entry.description,
			}); err != nil {
				return err
			}
			if entry.email != nil {
				pendingEmails = append(pendingEmails, *entry.email)
			}
		}

		updated, err := store.GetUserByID(ctx, targetUserId)
		if err != nil {
			return err
		}

		updatedProfile := buildAdminUserProfile(updated)
		profile = &updatedProfile
		return nil
	})

	if updateErr != nil {
		action := actionUserUpdated
		description := "Failed to update user"

		var auditErr *auditableError
		if errors.As(updateErr, &auditErr) {
			action = auditErr.action
			description = auditErr.description
		}

		// The transaction is gone, so the failure is recorded through the
		// pooled repository rather than the rolled back one.
		if logErr := s.createAdminAuditLog(ctx, s.adminRepository, AdminAuditLogInput{
			ActorUserID:  actorId,
			Action:       action,
			TargetUserID: targetUserId,
			Outcome:      db.AdminAuditOutcomeTypeFailed,
			RequestID:    requestId,
			Description:  description,
		}); logErr != nil {
			return nil, errors.Join(updateErr, logErr)
		}

		return nil, updateErr
	}

	for _, email := range pendingEmails {
		s.sendUserEmail(ctx, targetEmail, targetUserId, email)
	}

	return profile, nil
}

// sendUserEmail renders and fires one admin-triggered email. The update it
// describes has already committed, so a failure here is logged and
// swallowed rather than surfaced to the caller.
func (s *AdminService) sendUserEmail(ctx context.Context, targetEmail, targetUserId string, email pendingUserEmail) {
	html, err := mailer.RenderEmail(mailer.EmailData{
		Title:      email.heading,
		Heading:    email.heading,
		Subheading: email.subheading,
		Rows:       email.rows,
	})
	if err != nil {
		slog.Error("render admin-triggered email failed", "error", err, "user_id", targetUserId)
		return
	}

	mailer.SendEmailAsync(
		[]string{targetEmail},
		email.heading,
		html,
		middleware.GetReqID(ctx),
		targetUserId,
	)
}

func (s *AdminService) GetAdminAuditLogs(ctx context.Context, filters AdminAuditLogFilters) ([]dto.AdminAuditLogResponse, int64, error) {
	// Ensure limit is a proper number. Shouldn't return too many items at once
	if filters.Limit <= 0 {
		filters.Limit = 25
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	actorName := strings.TrimSpace(filters.ActorName)
	total, err := s.adminRepository.CountAdminAuditLogs(ctx, pgtype.Text{
		String: actorName,
		Valid:  actorName != "",
	})

	rows, err := s.adminRepository.GetAdminAuditLogs(ctx, db.GetAdminAuditLogsParams{
		ActorName: pgtype.Text{
			String: actorName,
			Valid:  actorName != "",
		},
		Limit:  filters.Limit,
		Offset: filters.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	logs := make([]dto.AdminAuditLogResponse, 0, len(rows))
	for _, row := range rows {
		var targetUser *dto.AdminAuditLogActor
		if row.TargetID.Valid {
			targetUser = &dto.AdminAuditLogActor{
				ActorUserId:    row.TargetID.String(),
				ActorFullName:  row.TargetName.String,
				ActorAvatarURL: row.TargetAvatarUrl.String,
			}
		}

		logs = append(logs, dto.AdminAuditLogResponse{
			Actor: dto.AdminAuditLogActor{
				ActorUserId:    row.ActorID.String(),
				ActorFullName:  row.ActorName,
				ActorAvatarURL: row.ActorAvatarUrl.String,
			},
			OccuredAt:   row.OccurredAt.Time,
			Action:      row.Action,
			Description: util.TextPointer(row.Description),
			Outcome:     dto.AdminAuditLogOutcomeType(row.Outcome),
			RequestId:   row.RequestID,
			TargetUser:  targetUser,
		})
	}

	return logs, total, nil
}

func (s *AdminService) ExportAuditLogs(
	ctx context.Context,
	filters AdminAuditLogFilters,
	actorId string,
	requestId string,
) ([]dto.AdminAuditLogResponse, error) {
	logs, exportErr := s.getAdminAuditLogs(ctx, buildAdminAuditLogParams(filters))

	outcome := db.AdminAuditOutcomeTypeSuccess
	description := fmt.Sprintf("Exported %d audit logs", len(logs))

	if exportErr != nil {
		outcome = db.AdminAuditOutcomeTypeFailed
		description = "Failed to export audit logs"
	}

	auditErr := s.createAdminAuditLog(ctx, s.adminRepository, AdminAuditLogInput{
		ActorUserID: actorId,
		Action:      "audit_logs.exported",
		Outcome:     outcome,
		RequestID:   requestId,
		Description: description,
	})

	if auditErr != nil {
		if exportErr != nil {
			return nil, errors.Join(exportErr, auditErr)
		}

		return nil, auditErr
	}

	if exportErr != nil {
		return nil, exportErr
	}

	return logs, nil
}

/*
	Private functions
*/

// applyUserUpdates performs every requested change and returns the audit
// entries the successful ones earned.
func (s *AdminService) applyUserUpdates(
	ctx context.Context,
	store repository.AdminStore,
	user db.GetAdminUserByIDRow,
	req UpdateUserRequest,
) ([]pendingAuditLog, error) {
	entries := make([]pendingAuditLog, 0)

	nameEntries, err := s.applyFullNameUpdate(ctx, store, user, req)
	if err != nil {
		return nil, err
	}
	entries = append(entries, nameEntries...)

	studentEntries, err := s.applyStudentUpdate(ctx, store, user, req)
	if err != nil {
		return nil, err
	}
	entries = append(entries, studentEntries...)

	roleEntries, err := s.applyRoleUpdate(ctx, store, user, req)
	if err != nil {
		return nil, err
	}
	entries = append(entries, roleEntries...)

	groupEntries, err := s.applyGroupUpdates(ctx, store, user, req)
	if err != nil {
		return nil, err
	}
	entries = append(entries, groupEntries...)

	membershipEntries, err := s.applyMembershipUpdates(ctx, store, user, req)
	if err != nil {
		return nil, err
	}
	entries = append(entries, membershipEntries...)

	return entries, nil
}

func (s *AdminService) applyFullNameUpdate(ctx context.Context, store repository.AdminStore, user db.GetAdminUserByIDRow, req UpdateUserRequest) ([]pendingAuditLog, error) {
	if req.FullName == nil {
		return nil, nil
	}
	fullName := strings.TrimSpace(*req.FullName)
	if fullName == "" {
		return nil, auditable(actionFullNameUpdated, "Failed to update full name: full name is required", fmt.Errorf("%w: full name is required", ErrValidation))
	}
	if fullName == user.FullName {
		return nil, nil
	}
	if err := store.UpdateUserFullName(ctx, user.ID.String(), fullName); err != nil {
		return nil, auditable(actionFullNameUpdated, "Failed to update full name", err)
	}
	return []pendingAuditLog{{
		action:      actionFullNameUpdated,
		description: fmt.Sprintf("Updated full name from %q to %q", user.FullName, fullName),
		email:       userInfoUpdateEmail("Full name", user.FullName, fullName),
	}}, nil
}

func (s *AdminService) applyStudentUpdate(
	ctx context.Context,
	store repository.AdminStore,
	user db.GetAdminUserByIDRow,
	req UpdateUserRequest,
) ([]pendingAuditLog, error) {
	update, err := planStudentUpdate(user, req)
	if err != nil {
		return nil, auditable(update.action, "Failed to update student details: "+err.Error(), err)
	}
	if !update.apply {
		return nil, nil
	}

	studentID := update.studentID
	if update.generate {
		studentID, err = generateUnusedNonStudentID(ctx, store)
		if err != nil {
			return nil, auditable(update.action, "Failed to update student status", err)
		}
	}

	if err := store.UpdateUserStudentInfo(ctx, user.ID.String(), update.isStudent, studentID); err != nil {
		if isUniqueViolation(err) {
			err = fmt.Errorf("%w: student ID %s is already in use", ErrConflict, studentID)
		}
		return nil, auditable(update.action, "Failed to update student details", err)
	}

	currentStudentID := textOrEmpty(user.StudentID)
	description := fmt.Sprintf("Updated student ID from %s to %s", displayValue(currentStudentID), studentID)
	item := "Student ID"
	oldValue := displayValue(currentStudentID)
	newValue := studentID
	if update.action == actionStudentStatusUpdated {
		description = fmt.Sprintf(
			"Updated student status from %s to %s (student ID %s to %s)",
			studentStatusLabel(user.IsStudent),
			studentStatusLabel(update.isStudent),
			displayValue(currentStudentID),
			studentID,
		)
		item = "Student status"
		oldValue = studentStatusLabel(user.IsStudent)
		newValue = studentStatusLabel(update.isStudent)
	}

	return []pendingAuditLog{{
		action:      update.action,
		description: description,
		email:       userInfoUpdateEmail(item, oldValue, newValue),
	}}, nil
}

func (s *AdminService) applyRoleUpdate(
	ctx context.Context,
	store repository.AdminStore,
	user db.GetAdminUserByIDRow,
	req UpdateUserRequest,
) ([]pendingAuditLog, error) {
	role, apply, err := planRoleUpdate(user, req)
	if err != nil {
		return nil, auditable(actionRoleUpdated, "Failed to update role: "+err.Error(), err)
	}
	if !apply {
		return nil, nil
	}

	if err := store.UpdateUserRole(ctx, user.ID.String(), role); err != nil {
		return nil, auditable(actionRoleUpdated, "Failed to update role", err)
	}

	return []pendingAuditLog{{
		action:      actionRoleUpdated,
		description: fmt.Sprintf("Updated role from %s to %s", user.Role, role),
		email:       userInfoUpdateEmail("Role", string(user.Role), string(role)),
	}}, nil
}

// userInfoUpdateEmail builds the "admin updated user info" email content for
// a single changed field.
func userInfoUpdateEmail(item, oldValue, newValue string) *pendingUserEmail {
	return &pendingUserEmail{
		heading:    "Your account information was updated",
		subheading: fmt.Sprintf("An admin updated your %s.", strings.ToLower(item)),
		rows: mailer.NewRows(
			"Updated", item,
			"Previous value", displayValue(oldValue),
			"New value", newValue,
		),
	}
}

// cancellationEmail builds the "admin cancelled your membership" email content.
func cancellationEmail(tierTitle string, cancelledAt time.Time) *pendingUserEmail {
	return &pendingUserEmail{
		heading:    "Your membership was cancelled",
		subheading: "An admin cancelled your membership. Reach out to us at communications@ubcesports.ca if you think this was a mistake.",
		rows: mailer.NewRows(
			"Tier", tierTitle,
			"Cancelled on", formatVancouverDate(cancelledAt),
		),
	}
}

func (s *AdminService) applyGroupUpdates(
	ctx context.Context,
	store repository.AdminStore,
	user db.GetAdminUserByIDRow,
	req UpdateUserRequest,
) ([]pendingAuditLog, error) {
	current := make(map[db.GroupType]struct{}, len(user.Groups))
	for _, group := range user.Groups {
		current[db.GroupType(group)] = struct{}{}
	}

	entries := make([]pendingAuditLog, 0, len(req.GroupsAdd)+len(req.GroupsRemove))

	for _, group := range req.GroupsAdd {
		if !isValidGroup(group) {
			err := fmt.Errorf("%w: invalid group %q", ErrValidation, group)
			return nil, auditable(actionGroupAdded, "Failed to add group: "+err.Error(), err)
		}
		if _, ok := current[group]; ok {
			continue
		}

		if err := store.AddUserGroup(ctx, user.ID.String(), group); err != nil {
			return nil, auditable(actionGroupAdded, fmt.Sprintf("Failed to add group %s", group), err)
		}

		hasExecGroup, err := store.HasExecGroup(ctx, user.ID.String())
		if err != nil {
			return nil, auditable(actionGroupAdded, "Failed to add group: "+err.Error(), err)
		}

		hasExecProfile, err := store.HasExecProfile(ctx, user.ID.String())
		if err != nil {
			return nil, auditable(actionGroupAdded, "Failed to add group: "+err.Error(), err)
		}

		if hasExecGroup && !hasExecProfile {
			err := store.CreateExecProfile(ctx, user.ID.String(), pgtype.Text{
				String: "Executive",
				Valid:  true,
			}, pgtype.Int4{
				Int32: 0,
				Valid: true,
			}, db.NullExecDisplayGroupType{
				ExecDisplayGroupType: defaultDisplayGroupType[group],
				Valid:                true,
			})
			if err != nil {
				return nil, auditable(actionGroupAdded, "Failed to add group: "+err.Error(), err)
			}
		}

		current[group] = struct{}{}
		entries = append(entries, pendingAuditLog{
			action:      actionGroupAdded,
			description: fmt.Sprintf("Added group %s", group),
		})
	}

	for _, group := range req.GroupsRemove {
		// The member group is permanent, so removal requests for it are dropped.
		if group == db.GroupTypeMember {
			continue
		}
		if !isValidGroup(group) {
			err := fmt.Errorf("%w: invalid group %q", ErrValidation, group)
			return nil, auditable(actionGroupRemoved, "Failed to remove group: "+err.Error(), err)
		}
		if _, ok := current[group]; !ok {
			continue
		}

		if err := store.RemoveUserGroup(ctx, user.ID.String(), group); err != nil {
			return nil, auditable(actionGroupRemoved, fmt.Sprintf("Failed to remove group %s", group), err)
		}

		hasExecGroup, err := store.HasExecGroup(ctx, user.ID.String())
		if err != nil {
			return nil, auditable(actionGroupRemoved, "Failed to remove group: "+err.Error(), err)
		}

		hasExecProfile, err := store.HasExecProfile(ctx, user.ID.String())
		if err != nil {
			return nil, auditable(actionGroupRemoved, "Failed to remove group: "+err.Error(), err)
		}

		if !hasExecGroup && hasExecProfile {
			if err := store.RemoveExecProfile(ctx, user.ID.String()); err != nil {
				return nil, auditable(actionGroupRemoved, "Failed to remove group: "+err.Error(), err)
			}
		}

		delete(current, group)
		entries = append(entries, pendingAuditLog{
			action:      actionGroupRemoved,
			description: fmt.Sprintf("Removed group %s", group),
		})
	}

	return entries, nil
}

func (s *AdminService) applyMembershipUpdates(
	ctx context.Context,
	store repository.AdminStore,
	user db.GetAdminUserByIDRow,
	req UpdateUserRequest,
) ([]pendingAuditLog, error) {
	if req.CancelMembershipId == nil {
		return nil, nil
	}

	membershipID := strings.TrimSpace(
		*req.CancelMembershipId,
	)

	if membershipID == "" {
		err := fmt.Errorf(
			"%w: membership ID is required",
			ErrValidation,
		)
		return nil, auditable(
			actionMembershipCancelled,
			"Failed to cancel membership: membership ID is required",
			err,
		)
	}

	if _, err := util.GetValidatedUUID(membershipID); err != nil {
		validationErr := fmt.Errorf(
			"%w: invalid membership ID",
			ErrValidation,
		)
		return nil, auditable(
			actionMembershipCancelled,
			"Failed to cancel membership: invalid membership ID",
			validationErr,
		)
	}

	// Best-effort lookup for the email; the cancellation itself only
	// depends on the CancelActiveMembershipByUserIdAndMembershipId call below.
	cancelledTierTitle := "your membership"
	if memberships, err := store.GetUserMemberships(ctx, user.ID.String()); err != nil {
		slog.Error("get user memberships for cancellation email failed", "error", err, "user_id", user.ID.String())
	} else {
		for _, membership := range memberships {
			if membership.ID.String() == membershipID {
				cancelledTierTitle = membership.TierTitle
				break
			}
		}
	}

	cancelledAt := time.Now()
	cancelled, err := s.adminRepository.CancelActiveMembershipByUserIdAndMembershipId(
		ctx,
		user.ID.String(),
		membershipID,
		cancelledAt,
	)
	if err != nil {
		return nil, auditable(
			actionMembershipCancelled,
			"Failed to cancel membership",
			err,
		)
	}

	if !cancelled {
		err := fmt.Errorf(
			"%w: membership is not active or does not belong to this user",
			ErrValidation,
		)
		return nil, auditable(
			actionMembershipCancelled,
			"Failed to cancel membership: no matching active membership",
			err,
		)
	}

	return []pendingAuditLog{{
		action: actionMembershipCancelled,
		description: fmt.Sprintf(
			"Cancelled membership %s",
			membershipID,
		),
		email: cancellationEmail(cancelledTierTitle, cancelledAt),
	}}, nil
}

// studentUpdate is the resolved outcome of the student ID and student status
// rules for a single request.
type studentUpdate struct {
	apply     bool
	isStudent bool
	studentID string
	generate  bool // when true, studentID is a randomly generated non-student ID
	action    string
}

func planStudentUpdate(user db.GetAdminUserByIDRow, req UpdateUserRequest) (studentUpdate, error) {
	statusChanged := req.IsStudent != nil && *req.IsStudent != user.IsStudent

	// Becoming a student requires a real student ID to go with it.
	if statusChanged && *req.IsStudent {
		update := studentUpdate{action: actionStudentStatusUpdated}
		if req.StudentID == nil {
			return update, fmt.Errorf("%w: student ID is required when marking a user as a student", ErrValidation)
		}

		studentID := strings.TrimSpace(*req.StudentID)
		if !studentIDRegex.MatchString(studentID) {
			return update, fmt.Errorf("%w: student ID must be an 8 digit number", ErrValidation)
		}

		update.apply = true
		update.isStudent = true
		update.studentID = studentID
		return update, nil
	}

	// Dropping student status replaces the student ID with a generated one, so
	// supplying one alongside would be silently discarded.
	if statusChanged && !*req.IsStudent {
		update := studentUpdate{
			apply:     true,
			isStudent: false,
			generate:  true,
			action:    actionStudentStatusUpdated,
		}
		if req.StudentID != nil {
			return studentUpdate{action: actionStudentStatusUpdated},
				fmt.Errorf("%w: student ID cannot be set when marking a user as a non-student", ErrValidation)
		}
		return update, nil
	}

	if req.StudentID == nil {
		return studentUpdate{}, nil
	}

	update := studentUpdate{action: actionStudentIDUpdated}
	if !user.IsStudent {
		return update, fmt.Errorf("%w: student ID can only be edited for students", ErrValidation)
	}

	studentID := strings.TrimSpace(*req.StudentID)
	if !studentIDRegex.MatchString(studentID) {
		return update, fmt.Errorf("%w: student ID must be an 8 digit number", ErrValidation)
	}
	if studentID == textOrEmpty(user.StudentID) {
		return studentUpdate{}, nil
	}

	update.apply = true
	update.isStudent = true
	update.studentID = studentID
	return update, nil
}

// planRoleUpdate resolves the requested role. An absent role leaves the user
// untouched, while an empty one floors the user back to member.
func planRoleUpdate(user db.GetAdminUserByIDRow, req UpdateUserRequest) (db.RoleType, bool, error) {
	if req.Role == nil {
		return user.Role, false, nil
	}

	role := db.RoleType(strings.TrimSpace(string(*req.Role)))
	switch role {
	case "":
		role = db.RoleTypeMember
	case db.RoleTypeMember, db.RoleTypeAdmin:
	default:
		return user.Role, false, fmt.Errorf("%w: role must be either member or admin", ErrValidation)
	}

	if role == user.Role {
		return role, false, nil
	}
	return role, true, nil
}

func generateUnusedNonStudentID(ctx context.Context, store repository.AdminStore) (string, error) {
	for range maxNonStudentIDAttempts {
		candidate := util.GenerateNonStudentID()

		exists, err := store.StudentIDExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to generate an unused non-student ID after %d attempts", maxNonStudentIDAttempts)
}

func isValidGroup(group db.GroupType) bool {
	switch group {
	case db.GroupTypeMember,
		db.GroupTypeCompetitiveTeam,
		db.GroupTypeExecutive,
		db.GroupTypeDirector,
		db.GroupTypeBoard:
		return true
	default:
		return false
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode
}

func studentStatusLabel(isStudent bool) string {
	if isStudent {
		return "student"
	}
	return "non-student"
}

func textOrEmpty(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func displayValue(value string) string {
	if value == "" {
		return "none"
	}
	return value
}

func buildAdminUserProfile(row db.GetAdminUserByIDRow) dto.ProfileDTO {
	groups := make([]dto.GroupType, 0, len(row.Groups))
	for _, group := range row.Groups {
		groups = append(groups, dto.GroupType(group))
	}

	return dto.ProfileDTO{
		ID:                    row.ID.String(),
		Email:                 row.Email,
		StudentID:             util.TextPointer(row.StudentID),
		Role:                  dto.RoleType(row.Role),
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
		FullName:              row.FullName,
		EmailVerifiedAt:       util.TimestampPointer(row.EmailVerifiedAt),
		IsStudent:             row.IsStudent,
		OnboardingCompletedAt: util.TimestampPointer(row.OnboardingCompletedAt),
		AvatarURL:             util.TextPointer(row.AvatarUrl),
		Groups:                groups,
	}
}

func (s *AdminService) createAdminAuditLog(
	ctx context.Context,
	store repository.AdminStore,
	input AdminAuditLogInput,
) error {
	actorID, err := util.GetValidatedUUID(input.ActorUserID)
	if err != nil {
		return fmt.Errorf("invalid audit actor user ID: %w", err)
	}

	targetID := pgtype.UUID{}
	if input.TargetUserID != "" {
		targetID, err = util.GetValidatedUUID(input.TargetUserID)
		if err != nil {
			return fmt.Errorf("invalid audit target user ID: %w", err)
		}
	}

	action := strings.TrimSpace(input.Action)
	if action == "" {
		return fmt.Errorf("audit action is required")
	}
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		return fmt.Errorf("audit request ID is required")
	}

	switch input.Outcome {
	case db.AdminAuditOutcomeTypeSuccess,
		db.AdminAuditOutcomeTypeFailed,
		db.AdminAuditOutcomeTypeDenied:
	default:
		return fmt.Errorf("invalid audit outcome: %q", input.Outcome)
	}

	description := strings.TrimSpace(input.Description)
	return store.CreateAdminAuditLog(ctx, db.CreateAdminAuditLogParams{
		ActorUserID:  actorID,
		Action:       action,
		TargetUserID: targetID,
		Outcome:      input.Outcome,
		RequestID:    requestID,
		Description: pgtype.Text{
			String: description,
			Valid:  description != "",
		},
	})
}

func buildAdminQueryParams(filters AdminUserFilters) db.GetUsersAdminParams {
	isStudent := pgtype.Bool{}
	if filters.IsStudent != nil {
		isStudent = pgtype.Bool{
			Bool:  *filters.IsStudent,
			Valid: true,
		}
	}

	groups := normalizedStrings(filters.Groups)
	membershipTierIDs := normalizedStrings(filters.MembershipTierIDs)

	return db.GetUsersAdminParams{
		FullName: pgtype.Text{
			String: filters.FullName,
			Valid:  filters.FullName != "",
		},
		StudentID: pgtype.Text{
			String: filters.StudentID,
			Valid:  filters.StudentID != "",
		},
		Email: pgtype.Text{
			String: filters.Email,
			Valid:  filters.Email != "",
		},
		Role: db.NullRoleType{
			RoleType: db.RoleType(filters.Role),
			Valid:    filters.Role != "",
		},
		IsStudent:         isStudent,
		Groups:            groups,
		MembershipTierIds: membershipTierIDs,
	}
}

func normalizedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil
	}
	sort.Strings(result)
	return result
}

func (s *AdminService) getUsers(ctx context.Context, params db.GetUsersAdminParams) ([]dto.AdminUserDTO, error) {
	rows, err := s.adminRepository.GetUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	users := make([]dto.AdminUserDTO, 0, len(rows))
	for _, row := range rows {
		groups := make([]dto.GroupType, 0, len(row.Groups))
		for _, group := range row.Groups {
			groups = append(groups, dto.GroupType(group))
		}

		activeMemberships := make([]dto.AdminActiveMembershipSummary, 0, len(row.ActiveMembershipTierTitles))
		for _, tierTitle := range row.ActiveMembershipTierTitles {
			activeMemberships = append(activeMemberships, dto.AdminActiveMembershipSummary{
				TierTitle: tierTitle,
			})
		}

		users = append(users, dto.AdminUserDTO{ProfileDTO: dto.ProfileDTO{
			ID:                    row.ID.String(),
			Email:                 row.Email,
			StudentID:             util.TextPointer(row.StudentID),
			Role:                  dto.RoleType(row.Role),
			CreatedAt:             row.CreatedAt.Time,
			UpdatedAt:             row.UpdatedAt.Time,
			FullName:              row.FullName,
			EmailVerifiedAt:       util.TimestampPointer(row.EmailVerifiedAt),
			IsStudent:             row.IsStudent,
			OnboardingCompletedAt: util.TimestampPointer(row.OnboardingCompletedAt),
			AvatarURL:             util.TextPointer(row.AvatarUrl),
			Groups:                groups,
		}, ActiveMemberships: activeMemberships})
	}

	return users, nil
}

func buildAdminAuditLogParams(filters AdminAuditLogFilters) db.GetAdminAuditLogsParams {
	actorName := strings.TrimSpace(filters.ActorName)
	return db.GetAdminAuditLogsParams{
		ActorName: pgtype.Text{
			String: actorName,
			Valid:  actorName != "",
		},
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}
}

func (s *AdminService) getAdminAuditLogs(ctx context.Context, params db.GetAdminAuditLogsParams) ([]dto.AdminAuditLogResponse, error) {
	rows, err := s.adminRepository.GetAdminAuditLogs(ctx, params)

	if err != nil {
		return nil, err
	}

	logs := make([]dto.AdminAuditLogResponse, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, dto.AdminAuditLogResponse{
			Actor: dto.AdminAuditLogActor{
				ActorUserId:    row.ActorID.String(),
				ActorFullName:  row.ActorName,
				ActorAvatarURL: row.ActorAvatarUrl.String,
			},
			OccuredAt:   row.OccurredAt.Time,
			Action:      row.Action,
			Description: util.TextPointer(row.Description),
			Outcome:     dto.AdminAuditLogOutcomeType(row.Outcome),
			RequestId:   row.RequestID,
			TargetUser: &dto.AdminAuditLogActor{
				ActorUserId:    row.TargetID.String(),
				ActorFullName:  row.TargetName.String,
				ActorAvatarURL: row.TargetAvatarUrl.String,
			},
		})
	}

	return logs, nil
}

func toNullExecDisplayGroupType(value db.NullGroupType) (db.NullExecDisplayGroupType, error) {
	if !value.Valid {
		return db.NullExecDisplayGroupType{}, ErrValidation
	}

	execDisplayGroupType := defaultDisplayGroupType[value.GroupType]
	if execDisplayGroupType == "" {
		return db.NullExecDisplayGroupType{}, ErrValidation
	}

	return db.NullExecDisplayGroupType{
		ExecDisplayGroupType: defaultDisplayGroupType[value.GroupType],
		Valid:                true,
	}, nil
}
