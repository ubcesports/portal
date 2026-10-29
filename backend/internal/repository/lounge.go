package repository

import (
	"context"

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

func (r *LoungeRepository) CheckoutLoungeSession(ctx context.Context, id string, checkedOutBy *string) (db.CheckoutLoungeSessionRow, error) {
	pgID, err := util.GetValidatedUUID(id)
	if err != nil {
		return db.CheckoutLoungeSessionRow{}, err
	}

	var pgCheckedOutBy pgtype.UUID
	if checkedOutBy != nil {
		parsedUUID, err := util.GetValidatedUUID(*checkedOutBy)
		if err != nil {
			return db.CheckoutLoungeSessionRow{}, err
		}
		pgCheckedOutBy = parsedUUID
	}

	return r.store.CheckoutLoungeSession(ctx, db.CheckoutLoungeSessionParams{
		ID:           pgID,
		CheckedOutBy: pgCheckedOutBy,
	})
}
