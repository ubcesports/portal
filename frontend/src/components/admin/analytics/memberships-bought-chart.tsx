import { Loader2 } from "lucide-react";
import { SurfacePanel } from "@/components/surface-panel";
import type { Granularity, MembershipsBoughtPoint } from "@/lib/types/admin.analytics.types";
import { formatNumber } from "@/lib/utils/formatting";
import { AnalyticsBarChart } from "./analytics-bar-chart";

// Slot 1 of the validated dark categorical palette (blue).
const CHART_COLOR = "#3987e5";

type MembershipsBoughtChartProps = {
  data: MembershipsBoughtPoint[];
  granularity: Granularity;
  isLoading: boolean;
};

/** Displays completed purchase counts by period, with loading and empty states. */
export function MembershipsBoughtChart({
  data,
  granularity,
  isLoading,
}: MembershipsBoughtChartProps) {
  return (
    <SurfacePanel className="flex flex-col gap-3 p-5">
      <h2 className="text-sm font-semibold text-brand-text">Memberships bought</h2>
      {isLoading ? (
        <div className="flex h-72 items-center justify-center text-brand-text-muted">
          <Loader2 aria-hidden="true" className="size-5 animate-spin" />
        </div>
      ) : data.length === 0 ? (
        <div className="flex h-72 items-center justify-center text-sm text-brand-text-subtle">
          No completed purchases in this range.
        </div>
      ) : (
        <AnalyticsBarChart
          data={data.map((point) => ({ periodStart: point.period_start, value: point.count }))}
          granularity={granularity}
          color={CHART_COLOR}
          valueLabel="Memberships bought"
          formatValue={formatNumber}
          allowDecimals={false}
        />
      )}
    </SurfacePanel>
  );
}
