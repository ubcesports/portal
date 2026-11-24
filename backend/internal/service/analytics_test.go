package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/repository"
)

func TestExportCSVUsesSharedTimeBounds(t *testing.T) {
	for _, granularity := range []string{"week", "month", "year"} {
		for _, sinceInception := range []bool{false, true} {
			filters := AnalyticsFilters{Granularity: granularity, Periods: 3, SinceInception: sinceInception}
			t.Run(granularity+map[bool]string{false: "/last_n", true: "/inception"}[sinceInception], func(t *testing.T) {
				store := &analyticsBoundsDB{}
				service := NewAnalyticsService(repository.NewAnalyticsRepository(nil, db.New(store)))
				rows, err := service.ExportCSV(context.Background(), filters)
				if err != nil {
					t.Fatal(err)
				}
				if len(store.bounds) != 2 {
					t.Fatalf("expected two series queries, got %d", len(store.bounds))
				}
				if store.bounds[0] != store.bounds[1] {
					t.Fatalf("series query bounds differ: %v", store.bounds)
				}
				if !store.bounds[0][1].Valid || store.bounds[0][0].Valid == sinceInception {
					t.Fatalf("incorrect bound validity: %v", store.bounds[0])
				}
				if sinceInception {
					if len(rows) != 0 {
						t.Fatal("expected no inception buckets without data")
					}
					return
				}
				if len(rows) != filters.Periods {
					t.Fatalf("expected %d buckets, got %d", filters.Periods, len(rows))
				}
				if !rows[0].PeriodStart.Equal(store.bounds[0][0].Time) {
					t.Fatal("export must start at the query lower bound")
				}
				if !rows[len(rows)-1].PeriodEnd.After(store.bounds[0][1].Time) {
					t.Fatal("last export bucket must contain the query upper bound")
				}
			})
		}
	}
}

// Capture the generated SQL arguments through the real service and repository.
type analyticsBoundsDB struct {
	db.DBTX
	bounds [][2]pgtype.Timestamptz
}

func (s *analyticsBoundsDB) Query(_ context.Context, _ string, args ...any) (pgx.Rows, error) {
	s.bounds = append(s.bounds, [2]pgtype.Timestamptz{args[1].(pgtype.Timestamptz), args[2].(pgtype.Timestamptz)})
	return emptyAnalyticsRows{}, nil
}

type emptyAnalyticsRows struct{ pgx.Rows }

func (emptyAnalyticsRows) Close()     {}
func (emptyAnalyticsRows) Next() bool { return false }
func (emptyAnalyticsRows) Err() error { return nil }
