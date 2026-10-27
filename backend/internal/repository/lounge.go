package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/util"
)

type LoungeRepository struct {
	store *db.Queries
}

func NewLoungeRepository(store *db.Queries) *LoungeRepository {
	return &LoungeRepository{store: store}
}

type CreateLoungeSessionParams struct {
	StudentID    string
	MembershipID string
	StationID    string
	CheckedInBy  string
	Game         *string
}

func (r *LoungeRepository) GetEligibleMembershipByStudentID(
	ctx context.Context,
	studentID string,
) (db.GetLoungeEligibleMembershipByStudentIDRow, error) {
	row, err := r.store.GetLoungeEligibleMembershipByStudentID(ctx, pgtype.Text{
		String: studentID,
		Valid:  true,
	})
	if err != nil {
		return db.GetLoungeEligibleMembershipByStudentIDRow{}, fmt.Errorf("query lounge-eligible membership by student ID: %w", err)
	}
	return row, nil
}

func (r *LoungeRepository) HasActiveSessionForUser(ctx context.Context, userID string) (bool, error) {
	pgUserID, err := util.GetValidatedUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid lounge session user ID: %w", err)
	}
	_, err = r.store.GetActiveLoungeSessionByUserID(ctx, pgUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("query active lounge session by user ID: %w", err)
	}
	return true, nil
}

func (r *LoungeRepository) CreateLoungeSession(
	ctx context.Context,
	params CreateLoungeSessionParams,
) (db.CreateLoungeSessionRow, error) {
	membershipID, err := util.GetValidatedUUID(params.MembershipID)
	if err != nil {
		return db.CreateLoungeSessionRow{}, fmt.Errorf("invalid lounge membership ID: %w", err)
	}
	stationID, err := util.GetValidatedUUID(params.StationID)
	if err != nil {
		return db.CreateLoungeSessionRow{}, fmt.Errorf("invalid lounge station ID: %w", err)
	}
	checkedInBy, err := util.GetValidatedUUID(params.CheckedInBy)
	if err != nil {
		return db.CreateLoungeSessionRow{}, fmt.Errorf("invalid lounge check-in actor ID: %w", err)
	}

	game := pgtype.Text{}
	if params.Game != nil {
		game = pgtype.Text{String: *params.Game, Valid: true}
	}
	row, err := r.store.CreateLoungeSession(ctx, db.CreateLoungeSessionParams{
		StudentID: pgtype.Text{
			String: params.StudentID,
			Valid:  true,
		},
		MembershipID: membershipID,
		StationID:    stationID,
		CheckedInBy:  checkedInBy,
		Game:         game,
	})
	if err != nil {
		return db.CreateLoungeSessionRow{}, fmt.Errorf("insert lounge session: %w", err)
	}
	return row, nil
}

func (r *LoungeRepository) GetLoungeStations(ctx context.Context) ([]db.GetLoungeStationsRow, error) {
	return r.store.GetLoungeStations(ctx)
}

func (r *LoungeRepository) GetLoungeStationByID(ctx context.Context, id string) (db.LoungeStation, error) {
	pgID, err := util.GetValidatedUUID(id)
	if err != nil {
		return db.LoungeStation{}, err
	}
	return r.store.GetLoungeStationByID(ctx, pgID)
}

func (r *LoungeRepository) GetLoungeStationByIDWithActiveSession(ctx context.Context, id string) (db.GetLoungeStationByIDWithActiveSessionRow, error) {
	pgID, err := util.GetValidatedUUID(id)
	if err != nil {
		return db.GetLoungeStationByIDWithActiveSessionRow{}, err
	}
	return r.store.GetLoungeStationByIDWithActiveSession(ctx, pgID)
}

func (r *LoungeRepository) UpdateLoungeStation(ctx context.Context, id string, label *string, closed *bool) (db.LoungeStation, error) {
	pgID, err := util.GetValidatedUUID(id)
	if err != nil {
		return db.LoungeStation{}, err
	}

	var pgLabel pgtype.Text
	if label != nil {
		pgLabel = pgtype.Text{String: *label, Valid: true}
	}

	var pgClosed pgtype.Bool
	if closed != nil {
		pgClosed = pgtype.Bool{Bool: *closed, Valid: true}
	}

	return r.store.UpdateLoungeStation(ctx, db.UpdateLoungeStationParams{
		ID:     pgID,
		Label:  pgLabel,
		Closed: pgClosed,
	})
}
