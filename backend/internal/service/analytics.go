package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
)

const (
	minAnalyticsPeriods = 1
	maxAnalyticsPeriods = 520
)

// AnalyticsFilters describes the filter/time-frame combination an admin
// analytics request is scoped to. Granularity/Periods/SinceInception are
// only meaningful for the time-series and export endpoints; the summary
// endpoint ignores them since its stats are always "current"/"all time".
type AnalyticsFilters struct {
	ProgramName    *string
	TierIDs        []string
	IsStudent      *bool
	PurchaseType   *string // "new" | "upgrade", nil = both
	Granularity    string  // "week" | "month" | "year"
	Periods        int     // used unless SinceInception is true
	SinceInception bool
}

type AnalyticsExportRow struct {
	PeriodStart       time.Time
	PeriodEnd         time.Time
	MembershipsBought int64
	RevenueCents      int64
}

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

func (s *AnalyticsService) GetSummary(ctx context.Context, filters AnalyticsFilters) (*dto.AnalyticsSummaryDTO, error) {
	programName := optionalText(filters.ProgramName)
	isStudent := optionalBool(filters.IsStudent)
	purchaseType := optionalPurchaseType(filters.PurchaseType)

	activeMemberships, err := s.analyticsRepo.GetActiveMembershipsCount(ctx, db.GetActiveMembershipsCountParams{
		ProgramName: programName,
		TierIds:     filters.TierIDs,
		IsStudent:   isStudent,
	})
	if err != nil {
		return nil, err
	}

	revenueCents, err := s.analyticsRepo.GetAllTimeRevenueCents(ctx, db.GetAllTimeRevenueCentsParams{
		ProgramName:  programName,
		TierIds:      filters.TierIDs,
		IsStudent:    isStudent,
		PurchaseType: purchaseType,
	})
	if err != nil {
		return nil, err
	}

	uniqueMembers, err := s.analyticsRepo.GetUniqueMembersCount(ctx, db.GetUniqueMembersCountParams{
		ProgramName:  programName,
		TierIds:      filters.TierIDs,
		IsStudent:    isStudent,
		PurchaseType: purchaseType,
	})
	if err != nil {
		return nil, err
	}

	return &dto.AnalyticsSummaryDTO{
		ActiveMemberships:   activeMemberships,
		AllTimeRevenueCents: revenueCents,
		UniqueMembers:       uniqueMembers,
	}, nil
}

func (s *AnalyticsService) GetMembershipsBoughtOverTime(
	ctx context.Context,
	filters AnalyticsFilters,
) ([]dto.MembershipsBoughtPointDTO, error) {
	now := time.Now().UTC()
	fromDate := resolveFromDate(filters, now)

	rows, err := s.analyticsRepo.GetMembershipsBoughtOverTime(ctx, db.GetMembershipsBoughtOverTimeParams{
		Granularity: filters.Granularity,
		FromDate:    toNullableTimestamptz(fromDate),
		ToDate:      toTimestamptz(now),
		ProgramName: optionalText(filters.ProgramName),
		TierIds:     filters.TierIDs,
		IsStudent:   optionalBool(filters.IsStudent),
	})
	if err != nil {
		return nil, err
	}

	counts := make(map[int64]int64, len(rows))
	var earliest *time.Time
	for _, row := range rows {
		bucketTime := row.Bucket.Time.UTC()
		counts[bucketTime.Unix()] = row.Count
		if earliest == nil || bucketTime.Before(*earliest) {
			earliest = &bucketTime
		}
	}

	buckets := expectedBuckets(filters.Granularity, fromDate, earliest, now)
	points := make([]dto.MembershipsBoughtPointDTO, 0, len(buckets))
	for _, bucket := range buckets {
		points = append(points, dto.MembershipsBoughtPointDTO{
			PeriodStart: bucket,
			Count:       counts[bucket.Unix()],
		})
	}
	return points, nil
}

func (s *AnalyticsService) GetRevenueOverTime(
	ctx context.Context,
	filters AnalyticsFilters,
) ([]dto.RevenuePointDTO, error) {
	now := time.Now().UTC()
	fromDate := resolveFromDate(filters, now)

	rows, err := s.analyticsRepo.GetRevenueOverTime(ctx, db.GetRevenueOverTimeParams{
		Granularity:  filters.Granularity,
		FromDate:     toNullableTimestamptz(fromDate),
		ToDate:       toTimestamptz(now),
		ProgramName:  optionalText(filters.ProgramName),
		TierIds:      filters.TierIDs,
		IsStudent:    optionalBool(filters.IsStudent),
		PurchaseType: optionalPurchaseType(filters.PurchaseType),
	})
	if err != nil {
		return nil, err
	}

	revenue := make(map[int64]int64, len(rows))
	var earliest *time.Time
	for _, row := range rows {
		bucketTime := row.Bucket.Time.UTC()
		revenue[bucketTime.Unix()] = row.RevenueCents
		if earliest == nil || bucketTime.Before(*earliest) {
			earliest = &bucketTime
		}
	}

	buckets := expectedBuckets(filters.Granularity, fromDate, earliest, now)
	points := make([]dto.RevenuePointDTO, 0, len(buckets))
	for _, bucket := range buckets {
		points = append(points, dto.RevenuePointDTO{
			PeriodStart:  bucket,
			RevenueCents: revenue[bucket.Unix()],
		})
	}
	return points, nil
}

// ExportCSV zips the two time series together by period start so the CSV has
// one row per bucket with both the membership count and the revenue.
func (s *AnalyticsService) ExportCSV(ctx context.Context, filters AnalyticsFilters) ([]AnalyticsExportRow, error) {
	memberships, err := s.GetMembershipsBoughtOverTime(ctx, filters)
	if err != nil {
		return nil, err
	}
	revenue, err := s.GetRevenueOverTime(ctx, filters)
	if err != nil {
		return nil, err
	}

	revenueByStart := make(map[int64]int64, len(revenue))
	for _, point := range revenue {
		revenueByStart[point.PeriodStart.Unix()] = point.RevenueCents
	}

	rows := make([]AnalyticsExportRow, 0, len(memberships))
	for _, point := range memberships {
		rows = append(rows, AnalyticsExportRow{
			PeriodStart:       point.PeriodStart,
			PeriodEnd:         addBuckets(point.PeriodStart, filters.Granularity, 1),
			MembershipsBought: point.Count,
			RevenueCents:      revenueByStart[point.PeriodStart.Unix()],
		})
	}
	return rows, nil
}

/*
	Private helpers
*/

// resolveFromDate returns nil (no lower bound) for "since inception", or the
// start of the Nth-from-last bucket for "last N periods".
func resolveFromDate(filters AnalyticsFilters, now time.Time) *time.Time {
	if filters.SinceInception {
		return nil
	}
	periods := clampPeriods(filters.Periods)
	start := addBuckets(truncateToBucket(now, filters.Granularity), filters.Granularity, -(periods - 1))
	return &start
}

// expectedBuckets returns every bucket start that should appear on the
// chart, so the caller can zero-fill buckets the query returned no rows for.
// When fromDate is nil (since inception) the range starts at the earliest
// bucket actually seen in the data; with no data at all, there's nothing to
// plot.
func expectedBuckets(granularity string, fromDate, earliest *time.Time, now time.Time) []time.Time {
	start := fromDate
	if start == nil {
		start = earliest
	}
	if start == nil {
		return nil
	}

	end := truncateToBucket(now, granularity)
	var buckets []time.Time
	for b := *start; !b.After(end); b = addBuckets(b, granularity, 1) {
		buckets = append(buckets, b)
	}
	return buckets
}

// truncateToBucket floors a UTC time to the start of its granularity bucket.
// Weeks start Monday, matching Postgres's date_trunc('week', ...) default.
func truncateToBucket(t time.Time, granularity string) time.Time {
	t = t.UTC()
	switch granularity {
	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		d := t.AddDate(0, 0, -(weekday - 1))
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	case "year":
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default: // "month"
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
}

func addBuckets(t time.Time, granularity string, n int) time.Time {
	switch granularity {
	case "week":
		return t.AddDate(0, 0, 7*n)
	case "year":
		return t.AddDate(n, 0, 0)
	default: // "month"
		return t.AddDate(0, n, 0)
	}
}

func clampPeriods(periods int) int {
	if periods < minAnalyticsPeriods {
		return minAnalyticsPeriods
	}
	if periods > maxAnalyticsPeriods {
		return maxAnalyticsPeriods
	}
	return periods
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalBool(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func optionalPurchaseType(value *string) db.NullPurchaseType {
	if value == nil {
		return db.NullPurchaseType{}
	}
	return db.NullPurchaseType{PurchaseType: db.PurchaseType(*value), Valid: true}
}

func toNullableTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
