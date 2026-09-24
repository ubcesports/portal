import apiClient from "../client";
import type {
  AnalyticsFilters,
  AnalyticsSummary,
  MembershipsBoughtPoint,
  RevenuePoint,
} from "@/lib/types/admin.analytics.types";

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
