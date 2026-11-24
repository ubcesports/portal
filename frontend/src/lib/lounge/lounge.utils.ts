import type { LoungeStation, LoungeStationStatus } from "@/lib/types/lounge.types";

/** Short label drawn on a map tile: "PC 7" → "7", the check-in kiosk → "You". */
export function getStationTileLabel(station: LoungeStation) {
  if (station.station_type === "check_in") {
    return "You";
  }
  if (station.station_type === "pc") {
    return station.label.replace(/^PC\s*/i, "");
  }
  return station.label;
}

/** A station that still has someone on it shows as in use even if it was closed. */
export function getStationStatus(station: LoungeStation): LoungeStationStatus {
  if (station.in_use) {
    return "in_use";
  }
  return station.closed ? "closed" : "available";
}

export function isSelectableStation(station: LoungeStation) {
  return station.station_type !== "check_in";
}

export function getElapsedMinutes(since: string, now: number) {
  return Math.max(0, Math.floor((now - new Date(since).getTime()) / 60_000));
}

/** Formats elapsed minutes as "39m" or "1h 05m". */
export function formatElapsed(minutes: number) {
  if (minutes < 60) {
    return `${minutes}m`;
  }
  const hours = Math.floor(minutes / 60);
  const remainder = String(minutes % 60).padStart(2, "0");
  return `${hours}h ${remainder}m`;
}

export function formatClockTime(value: string) {
  return new Date(value).toLocaleTimeString("en-CA", { hour: "numeric", minute: "2-digit" });
}
