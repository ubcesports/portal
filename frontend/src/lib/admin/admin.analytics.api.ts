import apiClient from "../client";
import type {
  AnalyticsFilters,
  AnalyticsSummary,
  MembershipsBoughtPoint,
  RevenuePoint,
} from "@/lib/types/admin.analytics.types";

/**
 * Serializes analytics filters, preserving repeated tier IDs and explicit false student status.
 * When includeTimeFrame is true, adds granularity and either periods or since_inception.
 */
export function buildAnalyticsParams(
  filters: AnalyticsFilters,
  includeTimeFrame: boolean,
): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.programName) {
    params.set("program_name", filters.programName);
  }

  for (const tierId of filters.tierIds ?? []) {
    params.append("tier_id", tierId);
  }

  if (filters.isStudent !== undefined) {
    params.set("is_student", String(filters.isStudent));
  }

  if (filters.purchaseType) {
    params.set("purchase_type", filters.purchaseType);
  }

  if (includeTimeFrame) {
    params.set("granularity", filters.granularity);

    if (filters.mode === "since_inception") {
      params.set("since_inception", "true");
    } else {
      params.set("periods", String(filters.periods));
    }
  }

  return params;
}

/**
 * Fetches current/all-time summary metrics, omitting time-range parameters.
 * The optional signal can cancel the request; request failures propagate to the caller.
 */
export async function fetchAnalyticsSummary(
  filters: AnalyticsFilters,
  signal?: AbortSignal,
): Promise<AnalyticsSummary> {
  const response = await apiClient.get<AnalyticsSummary>("/admin/analytics/summary", {
    params: buildAnalyticsParams(filters, false),
    signal,
  });

  return response.data;
}

/**
 * Fetches completed purchase counts for the selected time range, defaulting null data to [].
 * The optional signal can cancel the request; request failures propagate to the caller.
 */
export async function fetchMembershipsBoughtOverTime(
  filters: AnalyticsFilters,
  signal?: AbortSignal,
): Promise<MembershipsBoughtPoint[]> {
  const response = await apiClient.get<MembershipsBoughtPoint[]>(
    "/admin/analytics/memberships-over-time",
    {
      params: buildAnalyticsParams(filters, true),
      signal,
    },
  );

  return response.data ?? [];
}

/**
 * Fetches revenue in cents for the selected time range, defaulting null data to [].
 * The optional signal can cancel the request; request failures propagate to the caller.
 */
export async function fetchRevenueOverTime(
  filters: AnalyticsFilters,
  signal?: AbortSignal,
): Promise<RevenuePoint[]> {
  const response = await apiClient.get<RevenuePoint[]>("/admin/analytics/revenue-over-time", {
    params: buildAnalyticsParams(filters, true),
    signal,
  });

  return response.data ?? [];
}

/**
 * Fetches a CSV blob containing membership counts and revenue for the selected filters.
 * The caller handles downloading it. The optional signal can cancel the request.
 */
export async function exportAnalyticsCSV(
  filters: AnalyticsFilters,
  signal?: AbortSignal,
): Promise<Blob> {
  const response = await apiClient.get<Blob>("/admin/analytics/export", {
    params: buildAnalyticsParams(filters, true),
    responseType: "blob",
    signal,
  });

  return response.data;
}
