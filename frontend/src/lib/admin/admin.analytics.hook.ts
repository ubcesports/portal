import { useQuery } from "@tanstack/react-query";
import {
  fetchAnalyticsSummary,
  fetchMembershipsBoughtOverTime,
  fetchRevenueOverTime,
} from "./admin.analytics.api";
import type { AnalyticsFilters } from "@/lib/types/admin.analytics.types";

/** Queries and caches summary metrics by filters; options.enabled defaults to true. */
export function useAnalyticsSummary(filters: AnalyticsFilters, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "analytics", "summary", filters],
    queryFn: ({ signal }) => fetchAnalyticsSummary(filters, signal),
    enabled: options?.enabled ?? true,
  });
}

/** Queries and caches purchase counts by filters; options.enabled defaults to true. */
export function useMembershipsBoughtOverTime(
  filters: AnalyticsFilters,
  options?: { enabled?: boolean },
) {
  return useQuery({
    queryKey: ["admin", "analytics", "memberships-over-time", filters],
    queryFn: ({ signal }) => fetchMembershipsBoughtOverTime(filters, signal),
    enabled: options?.enabled ?? true,
  });
}

/** Queries and caches revenue by filters; options.enabled defaults to true. */
export function useRevenueOverTime(filters: AnalyticsFilters, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ["admin", "analytics", "revenue-over-time", filters],
    queryFn: ({ signal }) => fetchRevenueOverTime(filters, signal),
    enabled: options?.enabled ?? true,
  });
}
