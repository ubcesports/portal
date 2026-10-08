"use client";

import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { Granularity } from "@/lib/types/admin.analytics.types";

type ChartPoint = {
  periodStart: string;
  value: number;
};

type AnalyticsBarChartProps = {
  data: ChartPoint[];
  granularity: Granularity;
  color: string;
  valueLabel: string;
  formatValue: (value: number) => string;
  allowDecimals?: boolean;
};

/** Formats a bucket start in the browser time zone using a label for the granularity. */
function formatBucketLabel(periodStart: string, granularity: Granularity): string {
  const date = new Date(periodStart);
  if (granularity === "year") {
    return date.toLocaleDateString("en", { year: "numeric", timeZone: "UTC" });
  }
  if (granularity === "month") {
    return date.toLocaleDateString("en", { month: "short", year: "2-digit", timeZone: "UTC" });
  }
  return date.toLocaleDateString("en", { month: "short", day: "numeric", timeZone: "UTC" });
}

/** Renders a responsive period bar chart with formatted axis values and tooltips. */
export function AnalyticsBarChart({
  data,
  granularity,
  color,
  valueLabel,
  formatValue,
  allowDecimals = true,
}: AnalyticsBarChartProps) {
  const chartData = data.map((point) => ({
    ...point,
    label: formatBucketLabel(point.periodStart, granularity),
  }));

  return (
    <div className="h-72 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={chartData} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
          <CartesianGrid stroke="var(--color-border)" strokeOpacity={0.25} vertical={false} />
          <XAxis
            dataKey="label"
            tick={{ fill: "var(--color-text-subtle)", fontSize: 12 }}
            axisLine={{ stroke: "var(--color-border)" }}
            tickLine={false}
            minTickGap={16}
          />
          <YAxis
            tick={{ fill: "var(--color-text-subtle)", fontSize: 12 }}
            axisLine={false}
            tickLine={false}
            width={56}
            allowDecimals={allowDecimals}
            tickFormatter={(value) => formatValue(Number(value))}
          />
          <Tooltip
            cursor={{ fill: "rgba(255,255,255,0.05)" }}
            content={({ active, payload }) => {
              if (!active || !payload?.length) {
                return null;
              }
              const point = payload[0];
              return (
                <div className="border border-brand-border bg-brand-surface px-3 py-2 text-sm shadow-lg">
                  <div className="text-brand-text-subtle">{String(point.payload.label)}</div>
                  <div className="font-semibold text-brand-text">
                    {valueLabel}: {formatValue(Number(point.value))}
                  </div>
                </div>
              );
            }}
          />
          <Bar dataKey="value" fill={color} radius={[4, 4, 0, 0]} maxBarSize={24} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
