import { Loader2 } from "lucide-react";
import { SurfacePanel } from "@/components/surface-panel";
import type { AnalyticsSummary } from "@/lib/types/admin.analytics.types";
import { formatCentsAsCurrency, formatNumber } from "@/lib/utils/formatting";

type StatTileProps = {
  label: string;
  value: string;
  isLoading: boolean;
};

/** Displays a labelled summary value or a loading indicator. */
function StatTile({ label, value, isLoading }: StatTileProps) {
  return (
    <SurfacePanel className="flex flex-col gap-2 p-5">
      <span className="text-sm text-brand-text-subtle">{label}</span>
      {isLoading ? (
        <Loader2 aria-hidden="true" className="size-5 animate-spin text-brand-text-muted" />
      ) : (
        <span className="text-2xl font-semibold text-brand-text">{value}</span>
      )}
    </SurfacePanel>
  );
}

type SummaryCardsProps = {
  summary?: AnalyticsSummary;
  isLoading: boolean;
};

/**
 * Displays active memberships, all-time revenue, and unique purchasers.
 * Missing summary values default to zero when loading has finished.
 */
export function SummaryCards({ summary, isLoading }: SummaryCardsProps) {
  return (
    <div className="grid gap-4 sm:grid-cols-3">
      <StatTile
        label="Active memberships"
        value={formatNumber(summary?.active_memberships ?? 0)}
        isLoading={isLoading}
      />
      <StatTile
        label="All-time revenue"
        value={formatCentsAsCurrency(summary?.all_time_revenue_cents ?? 0)}
        isLoading={isLoading}
      />
      <StatTile
        label="Unique members"
        value={formatNumber(summary?.unique_members ?? 0)}
        isLoading={isLoading}
      />
    </div>
  );
}
