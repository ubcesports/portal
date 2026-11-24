import { AlertTriangle, Loader2, RefreshCw } from "lucide-react";
import { ActionButton } from "@/components/action-button";
import { SurfacePanel } from "@/components/surface-panel";
import { LoungeMap } from "@/components/lounge/lounge-map";
import { LoungeStationPanel } from "@/components/lounge/lounge-station-panel";
import { useLoungeStations } from "@/lib/lounge/lounge.hook";
import { useNow } from "@/lib/use-now.hook";

const ELAPSED_TICK_MS = 15_000;

type LoungeMapSectionProps = {
  selectedStationId: string | null;
  onSelectStation: (stationId: string | null) => void;
};

/** The lounge floor plan plus details for whichever station is selected. */
export function LoungeMapSection({ selectedStationId, onSelectStation }: LoungeMapSectionProps) {
  const { data: stations, isPending, isError, isFetching, refetch } = useLoungeStations();
  const now = useNow(ELAPSED_TICK_MS);

  const selectedStation = stations?.find((station) => station.id === selectedStationId) ?? null;

  return (
    <section className="flex flex-col border border-brand-border bg-brand-surface/80 shadow-2xl shadow-black/25">
      <div className="flex items-center justify-between gap-3 border-b border-brand-border px-5 py-5 sm:px-6">
        <div>
          <h2 className="text-lg font-semibold text-brand-text">Lounge map</h2>
          <p className="mt-1 text-sm text-brand-text-subtle">
            Select a station to see who&apos;s on it or check them out.
          </p>
        </div>
        {isFetching && !isPending ? (
          <Loader2
            aria-label="Refreshing stations"
            className="size-4 shrink-0 animate-spin text-brand-text-subtle"
          />
        ) : null}
      </div>

      <div className="px-5 py-6 sm:px-6">
        {isPending ? (
          <div className="flex min-h-64 items-center justify-center gap-3 text-brand-text-muted">
            <Loader2 aria-hidden="true" className="size-5 animate-spin" />
            <span>Loading lounge stations</span>
          </div>
        ) : isError ? (
          <SurfacePanel
            role="alert"
            className="flex flex-col gap-4 border-red-400/35 bg-red-400/10 px-5 py-5 sm:flex-row sm:items-center sm:justify-between"
          >
            <div className="flex items-start gap-3">
              <AlertTriangle aria-hidden="true" className="mt-0.5 size-5 shrink-0 text-red-200" />
              <div>
                <h3 className="text-sm font-semibold text-red-100">Unable to load the lounge</h3>
                <p className="mt-1 text-sm leading-6 text-brand-text-muted">
                  We couldn&apos;t load the stations. Try again.
                </p>
              </div>
            </div>
            <ActionButton
              className="shrink-0 border-red-300/40 text-red-100 hover:border-red-200/60 hover:bg-red-300/10"
              onClick={() => refetch()}
              loading={isFetching}
              icon={<RefreshCw aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            >
              {isFetching ? "Retrying" : "Try again"}
            </ActionButton>
          </SurfacePanel>
        ) : stations.length === 0 ? (
          <p className="py-10 text-center text-sm text-brand-text-muted">
            No lounge stations have been set up yet.
          </p>
        ) : (
          <LoungeMap
            stations={stations}
            selectedStationId={selectedStationId}
            now={now}
            onSelectStation={onSelectStation}
          />
        )}
      </div>

      {stations && stations.length > 0 ? (
        <div className="border-t border-brand-border" aria-live="polite">
          <LoungeStationPanel station={selectedStation} now={now} />
        </div>
      ) : null}
    </section>
  );
}
