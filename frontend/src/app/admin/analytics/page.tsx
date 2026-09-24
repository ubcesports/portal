"use client";

import { useMemo, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { BasePage } from "@/components/layout/base-page";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import { useAdminMembershipTierOptions } from "@/lib/admin/admin.hook";
import {
  useAnalyticsSummary,
  useMembershipsBoughtOverTime,
  useRevenueOverTime,
} from "@/lib/admin/admin.analytics.hook";
import { downloadCSVBlob } from "@/lib/admin/admin.api";
import { exportAnalyticsCSV } from "@/lib/admin/admin.analytics.api";
import {
  DEFAULT_ANALYTICS_FILTERS,
  type AnalyticsFilters,
} from "@/lib/types/admin.analytics.types";
import { AnalyticsFiltersToolbar } from "@/components/admin/analytics/analytics-filters-toolbar";
import { SummaryCards } from "@/components/admin/analytics/summary-cards";
import { MembershipsBoughtChart } from "@/components/admin/analytics/memberships-bought-chart";
import { RevenueChart } from "@/components/admin/analytics/revenue-chart";

export default function AnalyticsPage() {
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const [filters, setFilters] = useState<AnalyticsFilters>(DEFAULT_ANALYTICS_FILTERS);

  const { data: tierOptions = [] } = useAdminMembershipTierOptions({ enabled: isAdmin });
  const programOptions = useMemo(
    () => Array.from(new Set(tierOptions.map((tier) => tier.program_name))).sort(),
    [tierOptions],
  );

  const { data: summary, isPending: isSummaryPending } = useAnalyticsSummary(filters, {
    enabled: isAdmin,
  });
  const { data: membershipsBought = [], isPending: isMembershipsPending } =
    useMembershipsBoughtOverTime(filters, { enabled: isAdmin });
  const { data: revenue = [], isPending: isRevenuePending } = useRevenueOverTime(filters, {
    enabled: isAdmin,
  });

  const { mutate: exportCSV, isPending: isExporting } = useMutation({
    mutationFn: () => exportAnalyticsCSV(filters),
    onSuccess: (blob) => {
      downloadCSVBlob(blob, "membership-analytics.csv");
      toast.success("Membership analytics exported");
    },
  });

  const handleFiltersChange = (patch: Partial<AnalyticsFilters>) => {
    setFilters((current) => ({ ...current, ...patch }));
  };

  return (
    <BasePage>
      <div className="flex flex-1 items-center py-6">
        <section className="mx-auto flex min-h-[85vh] w-full flex-col">
          {isProfilePending ? (
            <div className="flex items-center gap-3 text-brand-text-muted">
              <Loader2 aria-hidden="true" className="size-5 animate-spin" />
              <span>Loading analytics</span>
            </div>
          ) : (
            <div className="flex min-h-[85vh] flex-1 flex-col border border-brand-border bg-brand-surface/80 shadow-2xl shadow-black/25">
              <div className="shrink-0 border-b border-brand-border px-5 py-5 sm:px-6">
                <h1 className="text-lg font-semibold text-brand-text">Analytics</h1>
                <p className="mt-1 text-sm text-brand-text-subtle">
                  Track memberships sold and revenue collected over time.
                </p>
              </div>

              <div className="shrink-0">
                <AnalyticsFiltersToolbar
                  filters={filters}
                  programOptions={programOptions}
                  tierOptions={tierOptions}
                  isExporting={isExporting}
                  onChange={handleFiltersChange}
                  onReset={() => setFilters(DEFAULT_ANALYTICS_FILTERS)}
                  onExport={() => exportCSV()}
                />
              </div>

              <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-y-auto p-5 sm:p-6">
                <SummaryCards summary={summary} isLoading={isSummaryPending} />

                <div className="grid gap-4 xl:grid-cols-2">
                  <MembershipsBoughtChart
                    data={membershipsBought}
                    granularity={filters.granularity}
                    isLoading={isMembershipsPending}
                  />
                  <RevenueChart
                    data={revenue}
                    granularity={filters.granularity}
                    isLoading={isRevenuePending}
                  />
                </div>
              </div>
            </div>
          )}
        </section>
      </div>
    </BasePage>
  );
}
