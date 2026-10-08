package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ubcesports/memberships/internal/database/db"
)

type AnalyticsRepository struct {
	pool  *pgxpool.Pool
	store *db.Queries
}

// NewAnalyticsRepository creates a repository using the supplied pool and queries.
func NewAnalyticsRepository(pool *pgxpool.Pool, store *db.Queries) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool, store: store}
}

// GetActiveMembershipsCount counts matching, uncancelled memberships active now.
// Student status is taken from the current user record.
func (r *AnalyticsRepository) GetActiveMembershipsCount(
	ctx context.Context,
	params db.GetActiveMembershipsCountParams,
) (int64, error) {
	count, err := r.store.GetActiveMembershipsCount(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count active memberships: %w", err)
	}
	return count, nil
}

// GetAllTimeRevenueCents sums matching completed payments in cents across all time.
// Student status is evaluated at purchase time.
func (r *AnalyticsRepository) GetAllTimeRevenueCents(
	ctx context.Context,
	params db.GetAllTimeRevenueCentsParams,
) (int64, error) {
	cents, err := r.store.GetAllTimeRevenueCents(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("sum all time revenue: %w", err)
	}
	return cents, nil
}

// GetUniqueMembersCount counts distinct users with matching completed purchases.
func (r *AnalyticsRepository) GetUniqueMembersCount(
	ctx context.Context,
	params db.GetUniqueMembersCountParams,
) (int64, error) {
	count, err := r.store.GetUniqueMembersCount(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("count unique members: %w", err)
	}
	return count, nil
}

// GetMembershipsBoughtOverTime counts matching completed purchases by period.
// The lower date bound is inclusive, the upper bound is exclusive, and periods
// without purchases are omitted. Student status is evaluated at purchase time.
func (r *AnalyticsRepository) GetMembershipsBoughtOverTime(
	ctx context.Context,
	params db.GetMembershipsBoughtOverTimeParams,
) ([]db.GetMembershipsBoughtOverTimeRow, error) {
	rows, err := r.store.GetMembershipsBoughtOverTime(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query memberships bought over time: %w", err)
	}
	return rows, nil
}

// GetRevenueOverTime sums matching completed payments in cents by period.
// The lower date bound is inclusive, the upper bound is exclusive, and periods
// without payments are omitted. Student status is evaluated at purchase time.
func (r *AnalyticsRepository) GetRevenueOverTime(
	ctx context.Context,
	params db.GetRevenueOverTimeParams,
) ([]db.GetRevenueOverTimeRow, error) {
	rows, err := r.store.GetRevenueOverTime(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query revenue over time: %w", err)
	}
	return rows, nil
}
