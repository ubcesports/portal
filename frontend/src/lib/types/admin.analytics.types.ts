export type Granularity = "week" | "month" | "year";

export type TimeRangeMode = "last_n" | "since_inception";

export type PurchaseType = "new" | "upgrade";

export type AnalyticsFilters = {
  programName?: string;
  tierIds?: string[];
  isStudent?: boolean;
  purchaseType?: PurchaseType; // undefined = both
  granularity: Granularity;
  mode: TimeRangeMode;
  periods: number; // used when mode === "last_n"
};

export type AnalyticsSummary = {
  active_memberships: number;
  all_time_revenue_cents: number;
  unique_members: number;
};

export type MembershipsBoughtPoint = {
  period_start: string;
  count: number;
};

export type RevenuePoint = {
  period_start: string;
  revenue_cents: number;
};

export const DEFAULT_ANALYTICS_FILTERS: AnalyticsFilters = {
  granularity: "week",
  mode: "last_n",
  periods: 12,
};

export const GRANULARITY_OPTIONS: { value: Granularity; label: string }[] = [
  { value: "week", label: "Weekly" },
  { value: "month", label: "Monthly" },
  { value: "year", label: "Yearly" },
];

export const TIME_RANGE_MODE_OPTIONS: { value: TimeRangeMode; label: string }[] = [
  { value: "last_n", label: "Last N periods" },
  { value: "since_inception", label: "Since inception" },
];

export const PURCHASE_TYPE_OPTIONS: { value: PurchaseType; label: string }[] = [
  { value: "new", label: "New purchases" },
  { value: "upgrade", label: "Upgrades" },
];

export const MIN_PERIODS = 1;
export const MAX_PERIODS = 520;
