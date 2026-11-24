package dto

import "time"

type AnalyticsSummaryDTO struct {
	ActiveMemberships   int64 `json:"active_memberships"`
	AllTimeRevenueCents int64 `json:"all_time_revenue_cents"`
	UniqueMembers       int64 `json:"unique_members"`
}

type MembershipsBoughtPointDTO struct {
	PeriodStart time.Time `json:"period_start"`
	Count       int64     `json:"count"`
}

type RevenuePointDTO struct {
	PeriodStart  time.Time `json:"period_start"`
	RevenueCents int64     `json:"revenue_cents"`
}
