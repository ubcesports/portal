import { Download, Loader2 } from "lucide-react";
import { ActionButton } from "@/components/action-button";
import { ToolbarContainer } from "@/components/toolbar/toolbar-container";
import { ToolbarRow } from "@/components/toolbar/toolbar-row";
import { SelectField } from "@/components/toolbar/select-option";
import { ResetButton } from "@/components/toolbar/reset-button";
import { CheckboxFilter } from "@/components/toolbar/checkbox-filter";
import type { AdminMembershipTierOption } from "@/lib/types/admin.types";
import { IS_STUDENT_OPTIONS, type IsStudentFilter } from "@/lib/types/admin.types";
import {
  DEFAULT_ANALYTICS_FILTERS,
  GRANULARITY_OPTIONS,
  MAX_PERIODS,
  MIN_PERIODS,
  PURCHASE_TYPE_OPTIONS,
  TIME_RANGE_MODE_OPTIONS,
  type AnalyticsFilters,
  type PurchaseType,
  type TimeRangeMode,
} from "@/lib/types/admin.analytics.types";

type AnalyticsFiltersToolbarProps = {
  filters: AnalyticsFilters;
  programOptions: string[];
  tierOptions: AdminMembershipTierOption[];
  isExporting: boolean;
  onChange: (patch: Partial<AnalyticsFilters>) => void;
  onReset: () => void;
  onExport: () => void;
};

function getIsStudentFilterValue(filters: AnalyticsFilters): IsStudentFilter {
  if (filters.isStudent === true) {
    return "yes";
  }
  if (filters.isStudent === false) {
    return "no";
  }
  return "all";
}

function hasActiveFilters(filters: AnalyticsFilters) {
  return (
    filters.programName !== undefined ||
    (filters.tierIds?.length ?? 0) > 0 ||
    filters.isStudent !== undefined ||
    filters.purchaseType !== undefined ||
    filters.granularity !== DEFAULT_ANALYTICS_FILTERS.granularity ||
    filters.mode !== DEFAULT_ANALYTICS_FILTERS.mode ||
    filters.periods !== DEFAULT_ANALYTICS_FILTERS.periods
  );
}

export function AnalyticsFiltersToolbar({
  filters,
  programOptions,
  tierOptions,
  isExporting,
  onChange,
  onReset,
  onExport,
}: AnalyticsFiltersToolbarProps) {
  const tierOptionsForProgram = filters.programName
    ? tierOptions.filter((tier) => tier.program_name === filters.programName)
    : tierOptions;

  return (
    <ToolbarContainer>
      <ToolbarRow>
        <SelectField
          label="Program"
          value={filters.programName ?? ""}
          onChange={(value) =>
            onChange({ programName: value || undefined, tierIds: undefined })
          }
          options={programOptions.map((program) => ({ value: program, label: program }))}
          allLabel="All programs"
          ariaLabel="Filter by program"
        />
        <SelectField
          label="Student status"
          value={getIsStudentFilterValue(filters)}
          onChange={(value) =>
            onChange({ isStudent: value === "all" ? undefined : value === "yes" })
          }
          options={IS_STUDENT_OPTIONS}
          allLabel="All"
          allValue="all"
          ariaLabel="Filter by student status"
        />
        <SelectField
          label="Purchase type (revenue only)"
          value={filters.purchaseType ?? ""}
          onChange={(value) => onChange({ purchaseType: (value || undefined) as PurchaseType })}
          options={PURCHASE_TYPE_OPTIONS}
          allLabel="Both"
          ariaLabel="Filter by purchase type"
        />
      </ToolbarRow>

      <CheckboxFilter
        label="Membership tiers"
        options={tierOptionsForProgram.map((tier) => ({
          value: tier.id,
          label: tier.title,
          description: tier.program_name,
        }))}
        selectedValues={filters.tierIds ?? []}
        onChange={(tierIds) => onChange({ tierIds: tierIds.length > 0 ? tierIds : undefined })}
        helpText="Matches any selected tier. Leave empty to include every tier."
      />

      <ToolbarRow>
        <SelectField
          label="Granularity"
          value={filters.granularity}
          onChange={(value) => onChange({ granularity: value as AnalyticsFilters["granularity"] })}
          options={GRANULARITY_OPTIONS}
          ariaLabel="Chart granularity"
        />
        <SelectField
          label="Time range"
          value={filters.mode}
          onChange={(value) => onChange({ mode: value as TimeRangeMode })}
          options={TIME_RANGE_MODE_OPTIONS}
          ariaLabel="Time range mode"
        />
        {filters.mode === "last_n" && (
          <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
            <span>Number of periods</span>
            <input
              type="number"
              min={MIN_PERIODS}
              max={MAX_PERIODS}
              value={filters.periods}
              onChange={(event) => {
                const parsed = Number(event.target.value);
                if (Number.isFinite(parsed)) {
                  onChange({ periods: Math.min(MAX_PERIODS, Math.max(MIN_PERIODS, parsed)) });
                }
              }}
              className="h-10 w-32 border border-brand-border bg-brand-surface px-3 text-sm text-brand-text"
              aria-label="Number of periods"
            />
          </label>
        )}
      </ToolbarRow>

      <div className="flex flex-wrap justify-end gap-2 border-t border-brand-border/70 pt-4">
        <ResetButton label="Reset Filters" onClick={onReset} disabled={!hasActiveFilters(filters)} />
        <ActionButton
          onClick={onExport}
          disabled={isExporting}
          loading={isExporting}
          icon={<Download aria-hidden="true" className="size-4" />}
          loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
        >
          {isExporting ? "Exporting" : "Export CSV"}
        </ActionButton>
      </div>
    </ToolbarContainer>
  );
}
