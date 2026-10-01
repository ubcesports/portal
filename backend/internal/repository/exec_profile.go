package repository

import (
	"context"

	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/util"
)

type ExecProfileRepository struct {
	store *db.Queries
}

func NewExecProfileRepository(store *db.Queries) *ExecProfileRepository {
	return &ExecProfileRepository{store: store}
}

func (r *ExecProfileRepository) GetExecProfiles(ctx context.Context) ([]db.GetExecProfilesRow, error) {
	return r.store.GetExecProfiles(ctx)
}

func (r *ExecProfileRepository) UpdateExecProfileByUserID(ctx context.Context, userId string, title string, displayOrder int32, displayGroup db.GroupType) (db.GetExecProfileByUserIDRow, error) {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, err
	}

	err = r.store.UpdateExecProfileTitle(ctx, db.UpdateExecProfileTitleParams{
		UserID: pgUserId,
		Title:  title,
	})

	updated_profile, err := r.store.GetExecProfileByUserID(ctx, pgUserId)
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, err
	}

	return updated_profile, nil
}

func (r *ExecProfileRepository) UpdateExecSocialLink(ctx context.Context, userId string, platform db.ExecSocialPlatformType, url string) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	return r.store.UpdateExecSocialLink(ctx, db.UpdateExecSocialLinkParams{
		UserID:   pgUserId,
		Platform: platform,
		Url:      url,
	})
}

func (r *ExecProfileRepository) DeleteExecSocialLink(ctx context.Context, userId string, platform db.ExecSocialPlatformType) error {
	pgUserId, err := util.GetValidatedUUID(userId)
	if err != nil {
		return err
	}

	return r.store.DeleteExecSocialLink(ctx, db.DeleteExecSocialLinkParams{
		UserID:   pgUserId,
		Platform: platform,
	})
}
