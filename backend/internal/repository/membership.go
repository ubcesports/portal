package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/util"
)

type CreatePendingTransactionParams struct {
	UserId            string
	TierId            string
	GroupAtPurchase   dto.GroupType
	StudentAtPurchase bool
	PurchaseType      dto.PurchaseType
	PaymentMethod     dto.PaymentMethodType
}

type CreateMembershipParams struct {
	UserId    string
	TierId    string
	StartedAt time.Time
	ExpiresAt time.Time
}

type CreateMembershipAuditLogParams struct {
	ActorUserId  string
	TargetUserId string
	Action       string
	Outcome      db.AdminAuditOutcomeType
	RequestId    string
	Description  string
}

type CompleteTransactionParams struct {
	TransactionId         string
	MembershipId          string
	StripePaymentIntentId string
	AmountPaidCents       int64
}

type MembershipRepository struct {
	pool  *pgxpool.Pool
	store *db.Queries
}

func NewMembershipRepository(pool *pgxpool.Pool, store *db.Queries) *MembershipRepository {
	return &MembershipRepository{pool: pool, store: store}
}

func (r *MembershipRepository) GetActiveTiersWithPrices(
	ctx context.Context,
) ([]dto.MembershipTierDTO, error) {
	rows, err := r.store.GetActiveTiersWithPrices(ctx)
	if err != nil {
		return nil, err
	}

	tiers := make([]dto.MembershipTierDTO, 0, len(rows))
	indexByTierID := make(map[string]int)

	for _, row := range rows {
		tierID := row.ID.String()

		var studentRequired *bool
		if row.IsStudentRequired.Valid {
			value := row.IsStudentRequired.Bool
			studentRequired = &value
		}

		price := dto.MembershipTierPriceDTO{
			Price:             float64(row.PriceInCents.Int64) / 100,
			PriceId:           row.StripePriceID.String,
			IsStudentRequired: studentRequired,
		}

		if index, exists := indexByTierID[tierID]; exists {
			tiers[index].Prices = append(
				tiers[index].Prices,
				price,
			)
			continue
		}

		indexByTierID[tierID] = len(tiers)

		tiers = append(tiers, dto.MembershipTierDTO{
			ID:          tierID,
			Title:       row.Title,
			Description: row.Description.String,
			Slug:        row.Slug.String,
			ProductId:   row.StripeProductID.String,
			Benefits:    row.Benefits,
			Limitations: row.Limitations,
			Prices:      []dto.MembershipTierPriceDTO{price},
			ProgramId:   row.ProgramID.String(),
			ProgramName: row.ProgramName,
			ExpirationType: dto.MembershipExpirationType(
				row.ExpirationType,
			),
			RequiredGroup: dto.GroupType(
				row.RequiredGroup.GroupType,
			),
		})
	}

	return tiers, nil
}

func (r *MembershipRepository) GetPublicTiersAndPrices(ctx context.Context) ([]db.GetPublicTiersAndPricesRow, error) {
	return r.store.GetPublicTiersAndPrices(ctx)
}

func (r *MembershipRepository) GetCurrentMembershipsWithTransactions(ctx context.Context, userId string) ([]dto.MembershipDTO, error) {
	var pgUserId pgtype.UUID

	if err := pgUserId.Scan(userId); err != nil {
		return []dto.MembershipDTO{}, err
	}

	memberships, err := r.store.GetCurrentMembershipsWithTransactions(ctx, pgUserId)
	if err != nil {
		return nil, err
	}

	returnMemberships := make([]dto.MembershipDTO, len(memberships))

	for i, m := range memberships {
		returnMemberships[i] = dto.MembershipDTO{
			ID:          m.ID.String(),
			TierId:      m.TierID.String(),
			TierTitle:   m.TierTitle,
			StartedAt:   m.StartedAt.Time,
			ExpiresAt:   m.ExpiresAt.Time,
			CancelledAt: util.TimestampPointer(m.CancelledAt),
			ProgramName: m.ProgramName,
			ProgramId:   m.ProgramID.String(),
			Slug:        m.Slug.String,
			Transaction: dto.TransactionDTO{
				ID:                    m.TransactionID.String(),
				AmountPaid:            fmt.Sprintf("%.2f", float64(m.AmountPaidCents.Int64)/100),
				Status:                dto.TransactionStatusType(m.Status),
				GroupAtPurchase:       dto.GroupType(m.GroupAtPurchase.GroupType),
				StudentAtPurchase:     m.StudentAtPurchase.Bool,
				StripePaymentIntentId: m.StripePaymentIntentID.String,
				PurchaseType:          dto.PurchaseType(m.PurchaseType.PurchaseType),
				AmountPaidCents:       m.AmountPaidCents.Int64,
			},
		}
	}

	return returnMemberships, nil
}

func (r *MembershipRepository) GetAllMembershipsWithTransactions(ctx context.Context, userId string) ([]db.GetAllMembershipsWithTransactionsRow, error) {
	var pgUserId pgtype.UUID

	if err := pgUserId.Scan(userId); err != nil {
		return []db.GetAllMembershipsWithTransactionsRow{}, err
	}

	return r.store.GetAllMembershipsWithTransactions(ctx, pgUserId)
}

func (r *MembershipRepository) GetTierByTierId(ctx context.Context, tierId string) (db.GetTierByTierIdRow, error) {
	var pgTierId pgtype.UUID

	if err := pgTierId.Scan(tierId); err != nil {
		return db.GetTierByTierIdRow{}, err
	}

	return r.store.GetTierByTierId(ctx, pgTierId)
}

func (r *MembershipRepository) GetPendingTransactionForUpdate(ctx context.Context, userId string) (db.GetPendingTransactionForUpdateRow, error) {
	var pgUserId pgtype.UUID

	if err := pgUserId.Scan(userId); err != nil {
		return db.GetPendingTransactionForUpdateRow{}, err
	}

	return r.store.GetPendingTransactionForUpdate(ctx, pgUserId)
}

func (r *MembershipRepository) ExpirePendingTransactionById(ctx context.Context, transactionId string) error {
	var pgTransactionId pgtype.UUID

	if err := pgTransactionId.Scan(transactionId); err != nil {
		return err
	}

	return r.store.ExpirePendingTransactionById(ctx, pgTransactionId)
}

func (r *MembershipRepository) PutStripeCheckoutSessionId(
	ctx context.Context,
	transactionId string,
	stripeCheckoutSessionId string,
) (int64, error) {
	var pgTransactionId pgtype.UUID
	if err := pgTransactionId.Scan(transactionId); err != nil {
		return 0, err
	}

	return r.store.PutStripeCheckoutSessionId(ctx, db.PutStripeCheckoutSessionIdParams{
		ID: pgTransactionId,
		StripeCheckoutSessionID: pgtype.Text{
			String: stripeCheckoutSessionId,
			Valid:  true,
		},
	})
}

func (r *MembershipRepository) UpdateTransactionStatusById(ctx context.Context, transactionId string, status dto.TransactionStatusType) error {
	var pgTransactionId pgtype.UUID

	if err := pgTransactionId.Scan(transactionId); err != nil {
		return err
	}

	return r.store.UpdateTransactionStatusById(ctx, db.UpdateTransactionStatusByIdParams{
		ID:     pgTransactionId,
		Status: db.TransactionStatusType(status),
	})
}

func (r *MembershipRepository) UpdatePendingTransactionStatusByCheckoutId(ctx context.Context, checkoutId string, status dto.TransactionStatusType) error {
	return r.store.UpdatePendingTransactionStatusByCheckoutId(ctx, db.UpdatePendingTransactionStatusByCheckoutIdParams{
		StripeCheckoutSessionID: pgtype.Text{
			String: checkoutId,
			Valid:  true,
		},
		Status: db.TransactionStatusType(status),
	})
}

func (r *MembershipRepository) CreatePendingTransaction(ctx context.Context, params CreatePendingTransactionParams) (string, error) {
	var userID pgtype.UUID
	if err := userID.Scan(params.UserId); err != nil {
		return "", err
	}
	var tierID pgtype.UUID
	if err := tierID.Scan(params.TierId); err != nil {
		return "", err
	}

	dbParams := db.CreatePendingTransactionParams{
		UserID: userID,
		TierID: tierID,
		GroupAtPurchase: db.NullGroupType{
			GroupType: db.GroupType(params.GroupAtPurchase),
			Valid:     true,
		},
		StudentAtPurchase: pgtype.Bool{
			Bool:  params.StudentAtPurchase,
			Valid: true,
		},
		PurchaseType: db.NullPurchaseType{
			PurchaseType: db.PurchaseType(params.PurchaseType),
			Valid:        true,
		},
		PaymentMethod: db.PaymentMethodType(params.PaymentMethod),
	}

	id, err := r.store.CreatePendingTransaction(ctx, dbParams)
	if err != nil {
		return "", err
	}

	return uuid.UUID(id.Bytes).String(), nil
}

func (r *MembershipRepository) GetTransactionByCheckoutSessionIdForUpdate(ctx context.Context, checkoutId string) (db.GetTransactionByCheckoutSessionIdForUpdateRow, error) {
	return r.store.GetTransactionByCheckoutSessionIdForUpdate(ctx, pgtype.Text{
		String: checkoutId,
		Valid:  true,
	})
}

func (r *MembershipRepository) GetTransactionByTransactionIdForUpdate(ctx context.Context, transactionId string) (db.GetTransactionByTransactionIdForUpdateRow, error) {
	var pgTransactionId pgtype.UUID
	if err := pgTransactionId.Scan(transactionId); err != nil {
		return db.GetTransactionByTransactionIdForUpdateRow{}, err
	}

	return r.store.GetTransactionByTransactionIdForUpdate(ctx, pgTransactionId)
}

func (r *MembershipRepository) CreateMembership(ctx context.Context, params CreateMembershipParams) (string, error) {
	var userID pgtype.UUID
	if err := userID.Scan(params.UserId); err != nil {
		return "", err
	}
	var tierID pgtype.UUID
	if err := tierID.Scan(params.TierId); err != nil {
		return "", err
	}

	id, err := r.store.CreateMembership(ctx, db.CreateMembershipParams{
		UserID: userID,
		TierID: tierID,
		StartedAt: pgtype.Timestamptz{
			Time:  params.StartedAt,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  params.ExpiresAt,
			Valid: true,
		},
	})
	if err != nil {
		return "", err
	}

	return uuid.UUID(id.Bytes).String(), nil
}

func (r *MembershipRepository) CompleteTransaction(ctx context.Context, params CompleteTransactionParams) error {
	var transactionID pgtype.UUID
	if err := transactionID.Scan(params.TransactionId); err != nil {
		return err
	}
	var membershipID pgtype.UUID
	if err := membershipID.Scan(params.MembershipId); err != nil {
		return err
	}

	return r.store.CompleteTransaction(ctx, db.CompleteTransactionParams{
		ID:           transactionID,
		MembershipID: membershipID,
		StripePaymentIntentID: pgtype.Text{
			String: params.StripePaymentIntentId,
			Valid:  params.StripePaymentIntentId != "",
		},
		AmountPaidCents: pgtype.Int8{
			Int64: params.AmountPaidCents,
			Valid: true,
		},
	})
}

func (r *MembershipRepository) CreateMembershipAuditLog(ctx context.Context, params CreateMembershipAuditLogParams) error {
	actorUserId, err := util.GetValidatedUUID(params.ActorUserId)
	if err != nil {
		return fmt.Errorf("invalid audit actor user ID: %w", err)
	}

	targetUserId, err := util.GetValidatedUUID(params.TargetUserId)
	if err != nil {
		return fmt.Errorf("invalid audit target user ID: %w", err)
	}

	return r.store.CreateAdminAuditLog(ctx, db.CreateAdminAuditLogParams{
		ActorUserID:  actorUserId,
		Action:       params.Action,
		TargetUserID: targetUserId,
		Outcome:      params.Outcome,
		RequestID:    params.RequestId,
		Description: pgtype.Text{
			String: params.Description,
			Valid:  params.Description != "",
		},
	})
}

func (r *MembershipRepository) CancelActiveMembershipsByUserIdAndProgramId(ctx context.Context, userId string, programId string, occurredAt time.Time) error {
	var pgUserID pgtype.UUID
	if err := pgUserID.Scan(userId); err != nil {
		return err
	}

	var pgProgramId pgtype.UUID
	if err := pgProgramId.Scan(programId); err != nil {
		return err
	}

	return r.store.CancelActiveMembershipsByUserIdAndProgramId(ctx, db.CancelActiveMembershipsByUserIdAndProgramIdParams{
		UserID:    pgUserID,
		ProgramID: pgProgramId,
		CancelledAt: pgtype.Timestamptz{
			Time:  occurredAt,
			Valid: true,
		},
	})
}

// GetActiveMembershipsExpiringOnDate returns active memberships whose
// expires_at falls on the given calendar date, in Vancouver time.
func (r *MembershipRepository) GetActiveMembershipsExpiringOnDate(ctx context.Context, date time.Time) ([]db.GetActiveMembershipsExpiringOnDateRow, error) {
	return r.store.GetActiveMembershipsExpiringOnDate(ctx, pgtype.Date{
		Time:  date,
		Valid: true,
	})
}

// executes fn within a database transaction.
//
// The callback receives a MembershipRepository whose operations are executed
// using the same transaction. If fn returns an error, the transaction is
// rolled back. Otherwise, the transaction is committed.
//
// This helper should be used when multiple repository operations must succeed
// or fail atomically.
func (r *MembershipRepository) WithTx(ctx context.Context, fn func(*MembershipRepository) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &MembershipRepository{
		pool:  r.pool,
		store: r.store.WithTx(tx),
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
