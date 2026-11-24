import type { CSSProperties } from "react";
import {
  formatElapsed,
  getElapsedMinutes,
  getStationStatus,
  getStationTileLabel,
  isSelectableStation,
} from "@/lib/lounge/lounge.utils";
import type { LoungeStation, LoungeStationStatus } from "@/lib/types/lounge.types";
import { cn } from "@/lib/utils/cn";

type LoungeMapProps = {
  stations: LoungeStation[];
  selectedStationId: string | null;
  now: number;
  onSelectStation: (stationId: string | null) => void;
};

const STATUS_TILE_STYLES: Record<LoungeStationStatus, string> = {
  available: "border-emerald-300/50 bg-emerald-500/30 text-emerald-50 hover:bg-emerald-500/45",
  in_use: "border-red-300/60 bg-red-500/45 text-red-50 hover:bg-red-500/60",
  closed: "border-brand-border bg-white/5 text-brand-text-subtle hover:bg-white/10",
};

const STATUS_LABELS: Record<LoungeStationStatus, string> = {
  available: "Available",
  in_use: "In use",
  closed: "Closed",
};

const LEGEND_SWATCHES: Record<LoungeStationStatus, string> = {
  available: "border-emerald-300/50 bg-emerald-500/30",
  in_use: "border-red-300/60 bg-red-500/45",
  closed: "border-brand-border bg-white/5",
};

function gridPlacement(station: LoungeStation): CSSProperties {
  return {
    gridRow: `${station.row + 1} / span ${station.height}`,
    gridColumn: `${station.col + 1} / span ${station.width}`,
  };
}

/** Draws the lounge floor plan from each station's grid position and size. */
export function LoungeMap({ stations, selectedStationId, now, onSelectStation }: LoungeMapProps) {
  const rowCount = Math.max(1, ...stations.map((station) => station.row + station.height));
  const columnCount = Math.max(1, ...stations.map((station) => station.col + station.width));

  return (
    <div className="flex flex-col items-center gap-5">
      <div
        className="grid gap-2 [--lounge-tile:3.25rem] sm:[--lounge-tile:4.5rem]"
        style={{
          gridTemplateColumns: `repeat(${columnCount}, var(--lounge-tile))`,
          gridTemplateRows: `repeat(${rowCount}, var(--lounge-tile))`,
        }}
      >
        {stations.map((station) => {
          const tileLabel = getStationTileLabel(station);

          if (!isSelectableStation(station)) {
            return (
              <div
                key={station.id}
                style={gridPlacement(station)}
                className="flex items-center justify-center border border-brand-text-muted/70 text-sm text-brand-text-muted"
                title={station.label}
              >
                {tileLabel}
              </div>
            );
          }

          const status = getStationStatus(station);
          const isSelected = station.id === selectedStationId;
          const elapsed =
            status === "in_use" && station.active_session_check_in_time
              ? formatElapsed(getElapsedMinutes(station.active_session_check_in_time, now))
              : null;

          return (
            <button
              key={station.id}
              type="button"
              style={gridPlacement(station)}
              aria-pressed={isSelected}
              aria-label={`${station.label}, ${STATUS_LABELS[status].toLowerCase()}${elapsed ? ` for ${elapsed}` : ""}`}
              onClick={() => onSelectStation(isSelected ? null : station.id)}
              className={cn(
                "flex cursor-pointer flex-col items-center justify-center border transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-text",
                STATUS_TILE_STYLES[status],
                status === "closed" &&
                  "bg-[repeating-linear-gradient(135deg,transparent_0_6px,rgb(255_255_255/0.05)_6px_12px)]",
                isSelected && "ring-2 ring-brand-text ring-offset-2 ring-offset-brand-surface",
              )}
            >
              <span className="text-base font-semibold sm:text-xl">{tileLabel}</span>
              {elapsed ? (
                <span className="text-[0.65rem] leading-tight opacity-90 sm:text-xs">
                  {elapsed}
                </span>
              ) : null}
            </button>
          );
        })}
      </div>

      <ul className="flex flex-wrap justify-center gap-x-5 gap-y-2 text-xs text-brand-text-subtle">
        {(Object.keys(STATUS_LABELS) as LoungeStationStatus[]).map((status) => (
          <li key={status} className="flex items-center gap-2">
            <span aria-hidden="true" className={cn("size-3 border", LEGEND_SWATCHES[status])} />
            {STATUS_LABELS[status]}
          </li>
        ))}
      </ul>
    </div>
  );
}
