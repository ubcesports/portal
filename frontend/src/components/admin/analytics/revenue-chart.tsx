import { Loader2 } from "lucide-react";
import { SurfacePanel } from "@/components/surface-panel";
import type { Granularity, RevenuePoint } from "@/lib/types/admin.analytics.types";
import { formatCentsAsCurrency } from "@/lib/utils/formatting";
import { AnalyticsBarChart } from "./analytics-bar-chart";

// Slot 3 of the validated dark categorical palette (aqua) - distinct from the
// memberships-bought chart's blue, per the categorical color-assignment rule.
const CHART_COLOR = "#199e70";

type RevenueChartProps = {
  data: RevenuePoint[];
  granularity: Granularity;
  isLoading: boolean;
};

export function RevenueChart({ data, granularity, isLoading }: RevenueChartProps) {
  return (
    <SurfacePanel className="flex flex-col gap-3 p-5">
      <h2 className="text-sm font-semibold text-brand-text">Revenue</h2>
      {isLoading ? (
        <div className="flex h-72 items-center justify-center text-brand-text-muted">
          <Loader2 aria-hidden="true" className="size-5 animate-spin" />
        </div>
      ) : data.length === 0 ? (
        <div className="flex h-72 items-center justify-center text-sm text-brand-text-subtle">
          No revenue in this range.
        </div>
      ) : (
        <AnalyticsBarChart
          data={data.map((point) => ({ periodStart: point.period_start, value: point.revenue_cents }))}
          granularity={granularity}
          color={CHART_COLOR}
          valueLabel="Revenue"
          formatValue={formatCentsAsCurrency}
        />
      )}
    </SurfacePanel>
  );
}
