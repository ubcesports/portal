import { AlertTriangle, Loader2, LogOut, MousePointerClick, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { DetailRow } from "@/components/detail-row";
import { StatusBadge } from "@/components/status-badge";
import { useCheckOutLoungeSession, useLoungeSession } from "@/lib/lounge/lounge.hook";
import {
  formatClockTime,
  formatElapsed,
  getElapsedMinutes,
  getStationStatus,
} from "@/lib/lounge/lounge.utils";
import type { LoungeStation, LoungeStationStatus } from "@/lib/types/lounge.types";

type LoungeStationPanelProps = {
  station: LoungeStation | null;
  now: number;
};

const STATION_TYPE_LABELS = {
  pc: "PC",
  cab: "Arcade cabinet",
  check_in: "Check-in counter",
} as const;

const STATUS_BADGES: Record<
  LoungeStationStatus,
  { label: string; tone: "success" | "warning" | "muted" }
> = {
  available: { label: "Available", tone: "success" },
  in_use: { label: "In use", tone: "warning" },
  closed: { label: "Closed", tone: "muted" },
};

/** Details for the selected station, including its active session if there is one. */
export function LoungeStationPanel({ station, now }: LoungeStationPanelProps) {
  if (!station) {
    return (
      <div className="flex items-center gap-3 px-5 py-6 text-sm text-brand-text-subtle sm:px-6">
        <MousePointerClick aria-hidden="true" className="size-4 shrink-0" />
        Select a PC or cabinet on the map to see its details.
      </div>
    );
  }

  const status = getStationStatus(station);
  const badge = STATUS_BADGES[status];

  return (
    <div>
      <div className="flex flex-wrap items-center gap-3 px-5 py-5 sm:px-6">
        <h2 className="text-xl font-semibold text-brand-text">{station.label}</h2>
        <StatusBadge tone={badge.tone}>{badge.label}</StatusBadge>
        {station.closed && status === "in_use" ? (
          <StatusBadge tone="muted">Closed after this session</StatusBadge>
        ) : null}
        <span className="text-sm text-brand-text-subtle">
          {STATION_TYPE_LABELS[station.station_type]}
        </span>
      </div>

      {station.active_session_id ? (
        <ActiveSessionDetails
          key={station.active_session_id}
          sessionId={station.active_session_id}
          stationLabel={station.label}
          now={now}
        />
      ) : (
        <p className="border-t border-brand-border px-5 py-4 text-sm text-brand-text-muted sm:px-6">
          {status === "closed"
            ? "This station is closed and can't be used right now."
            : "No one is using this station."}
        </p>
      )}
    </div>
  );
}

type ActiveSessionDetailsProps = {
  sessionId: string;
  stationLabel: string;
  now: number;
};

function ActiveSessionDetails({ sessionId, stationLabel, now }: ActiveSessionDetailsProps) {
  const { data: session, isPending, isError, isFetching, refetch } = useLoungeSession(sessionId);
  const { mutate: checkOut, isPending: isCheckingOut } = useCheckOutLoungeSession();

  const handleCheckOut = () => {
    checkOut(sessionId, {
      onSuccess: () => toast.success(`${stationLabel} checked out`),
    });
  };

  const checkOutButton = (
    <ActionButton
      className="border-red-300/50 text-red-100 hover:border-red-200/70 hover:bg-red-400/10"
      onClick={handleCheckOut}
      loading={isCheckingOut}
      icon={<LogOut aria-hidden="true" className="size-4" />}
      loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
    >
      {isCheckingOut ? "Checking out" : "Check out"}
    </ActionButton>
  );

  if (isPending) {
    return (
      <div className="flex items-center gap-3 border-t border-brand-border px-5 py-4 text-sm text-brand-text-muted sm:px-6">
        <Loader2 aria-hidden="true" className="size-4 animate-spin" />
        Loading session
      </div>
    );
  }

  if (isError) {
    // Staff can still free the station even if the member details failed to load.
    return (
      <div
        role="alert"
        className="flex flex-col gap-4 border-t border-brand-border px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6"
      >
        <div className="flex items-center gap-3 text-sm text-red-100">
          <AlertTriangle aria-hidden="true" className="size-4 shrink-0" />
          Unable to load this session&apos;s details.
        </div>
        <div className="flex flex-wrap gap-2">
          <ActionButton
            onClick={() => refetch()}
            loading={isFetching}
            icon={<RefreshCw aria-hidden="true" className="size-4" />}
            loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
          >
            Try again
          </ActionButton>
          {checkOutButton}
        </div>
      </div>
    );
  }

  const elapsed = formatElapsed(getElapsedMinutes(session.check_in_time, now));

  return (
    <div>
      <dl>
        <DetailRow label="Member">
          <span className="font-medium">{session.user_full_name}</span>
          {session.user_student_id ? (
            <span className="text-brand-text-subtle"> · {session.user_student_id}</span>
          ) : null}
        </DetailRow>
        <DetailRow label="Membership">{session.membership_tier_title ?? "None recorded"}</DetailRow>
        <DetailRow label="Game">{session.game ?? "Not specified"}</DetailRow>
        <DetailRow label="Started">
          {formatClockTime(session.check_in_time)}
          <span className="text-brand-text-subtle"> · {elapsed} ago</span>
        </DetailRow>
        <DetailRow label="Checked in by">{session.checked_in_by_name}</DetailRow>
      </dl>
      <div className="flex justify-end border-t border-brand-border px-5 py-4 sm:px-6">
        {checkOutButton}
      </div>
    </div>
  );
}
