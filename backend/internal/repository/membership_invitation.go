package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/util"
)

type MembershipInvitationRepository struct {
	pool  *pgxpool.Pool
	store *db.Queries
}

type MembershipInvitationRedemption struct {
	ID              string
	TierID          string
	AmountPaidCents int64
	PaymentMethod   dto.PaymentMethodType
	PurchasedAt     time.Time
	ProgramID       string
	ExpirationType  dto.MembershipExpirationType
}

type MembershipInvitationAuditLogParams struct {
	ActorUserID string
	Action      string
	Outcome     db.AdminAuditOutcomeType
	RequestID   string
	Description string
}

func NewMembershipInvitationRepository(
	pool *pgxpool.Pool,
	store *db.Queries,
) *MembershipInvitationRepository {
	return &MembershipInvitationRepository{pool: pool, store: store}
}

func (r *MembershipInvitationRepository) Create(
	ctx context.Context,
	request dto.CreateMembershipInvitationRequest,
	createdByUserID string,
) (*dto.MembershipInvitationDTO, error) {
	tierID, err := util.GetValidatedUUID(request.TierID)
	if err != nil {
		return nil, err
	}
	creatorID, err := util.GetValidatedUUID(createdByUserID)
	if err != nil {
		return nil, err
	}

	row, err := r.store.CreateMembershipInvitation(ctx, db.CreateMembershipInvitationParams{
		Email:           request.Email,
		TierID:          tierID,
		AmountPaidCents: request.AmountPaidCents,
		PaymentMethod:   db.PaymentMethodType(request.PaymentMethod),
		CreatedByUserID: creatorID,
	})
	if err != nil {
		return nil, err
	}

	return membershipInvitationDTO(
		row.ID,
		row.Email,
		row.TierID,
		row.TierTitle,
		row.ProgramName,
		row.AmountPaidCents,
		row.PaymentMethod,
		row.Done,
		row.CreatedByUserID,
		row.CreatedByName,
		row.InvitationSentAt,
		row.PurchasedAt,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *MembershipInvitationRepository) GetAll(ctx context.Context) ([]dto.MembershipInvitationDTO, error) {
	rows, err := r.store.GetMembershipInvitations(ctx)
	if err != nil {
		return nil, err
	}

	invitations := make([]dto.MembershipInvitationDTO, 0, len(rows))
	for _, row := range rows {
		invitations = append(invitations, *membershipInvitationDTO(
			row.ID,
			row.Email,
			row.TierID,
			row.TierTitle,
			row.ProgramName,
			row.AmountPaidCents,
			row.PaymentMethod,
			row.Done,
			row.CreatedByUserID,
			row.CreatedByName,
			row.InvitationSentAt,
			row.PurchasedAt,
			row.CreatedAt,
			row.UpdatedAt,
		))
	}
	return invitations, nil
}

func (r *MembershipInvitationRepository) GetByID(ctx context.Context, id string) (*dto.MembershipInvitationDTO, error) {
	invitationID, err := util.GetValidatedUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.store.GetMembershipInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	return membershipInvitationDTO(
		row.ID,
		row.Email,
		row.TierID,
		row.TierTitle,
		row.ProgramName,
		row.AmountPaidCents,
		row.PaymentMethod,
		row.Done,
		row.CreatedByUserID,
		row.CreatedByName,
		row.InvitationSentAt,
		row.PurchasedAt,
		row.CreatedAt,
		row.UpdatedAt,
	), nil
}

func (r *MembershipInvitationRepository) Delete(ctx context.Context, id string) error {
	invitationID, err := util.GetValidatedUUID(id)
	if err != nil {
		return err
	}

	rowsAffected, err := r.store.DeleteMembershipInvitation(ctx, invitationID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MembershipInvitationRepository) UserExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	return r.store.MembershipInvitationUserExistsByEmail(ctx, email)
}

func (r *MembershipInvitationRepository) GetByEmailForUpdate(
	ctx context.Context,
	email string,
) (*MembershipInvitationRedemption, error) {
	row, err := r.store.GetMembershipInvitationByEmailForUpdate(ctx, email)
	if err != nil {
		return nil, err
	}

	return &MembershipInvitationRedemption{
		ID:              row.ID.String(),
		TierID:          row.TierID.String(),
		AmountPaidCents: row.AmountPaidCents,
		PaymentMethod:   dto.PaymentMethodType(row.PaymentMethod),
		PurchasedAt:     row.PurchasedAt.Time,
		ProgramID:       row.ProgramID.String(),
		ExpirationType:  dto.MembershipExpirationType(row.ExpirationType),
	}, nil
}

func (r *MembershipInvitationRepository) MarkDone(ctx context.Context, id string) error {
	invitationID, err := util.GetValidatedUUID(id)
	if err != nil {
		return err
	}

	rowsAffected, err := r.store.MarkMembershipInvitationDone(ctx, invitationID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MembershipInvitationRepository) CreateAuditLog(
	ctx context.Context,
	params MembershipInvitationAuditLogParams,
) error {
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

// WithTx executes membership-invitation operations in one transaction. It is used
// to keep successful mutations and their audit entries atomic.
func (r *MembershipInvitationRepository) WithTx(
	ctx context.Context,
	fn func(*MembershipInvitationRepository) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &MembershipInvitationRepository{pool: r.pool, store: r.store.WithTx(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// WithRedemptionTx executes invitation and membership operations in one
// transaction so an entitlement is only marked done after it is fulfilled.
func (r *MembershipInvitationRepository) WithRedemptionTx(
	ctx context.Context,
	fn func(*MembershipInvitationRepository, *MembershipRepository) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txStore := r.store.WithTx(tx)
	txInvitationRepo := &MembershipInvitationRepository{pool: r.pool, store: txStore}
	txMembershipRepo := &MembershipRepository{pool: r.pool, store: txStore}
	if err := fn(txInvitationRepo, txMembershipRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func membershipInvitationDTO(
	id pgtype.UUID,
	email string,
	tierID pgtype.UUID,
	tierTitle string,
	programName string,
	amountPaidCents int64,
	paymentMethod db.PaymentMethodType,
	done bool,
	createdByUserID pgtype.UUID,
	createdByName string,
	invitationSentAt pgtype.Timestamptz,
	purchasedAt pgtype.Timestamptz,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
) *dto.MembershipInvitationDTO {
	var sentAt *time.Time
	if invitationSentAt.Valid {
		sentAt = &invitationSentAt.Time
	}

	return &dto.MembershipInvitationDTO{
		ID:               id.String(),
		Email:            email,
		TierID:           tierID.String(),
		TierTitle:        tierTitle,
		ProgramName:      programName,
		AmountPaidCents:  amountPaidCents,
		PaymentMethod:    dto.PaymentMethodType(paymentMethod),
		Done:             done,
		CreatedByUserID:  createdByUserID.String(),
		CreatedByName:    createdByName,
		InvitationSentAt: sentAt,
		PurchasedAt:      purchasedAt.Time,
		CreatedAt:        createdAt.Time,
		UpdatedAt:        updatedAt.Time,
	}
}
