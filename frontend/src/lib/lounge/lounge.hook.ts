import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { checkOutLoungeSession, fetchLoungeSession, fetchLoungeStations } from "./lounge.api";

const STATIONS_REFETCH_INTERVAL_MS = 30_000;

/**
 * Query keys shared by every lounge feature. Anything that starts or ends a
 * session (e.g. check-in) should invalidate `loungeQueryKeys.stations` so the
 * map picks up the change.
 */
export const loungeQueryKeys = {
  all: ["lounge"] as const,
  stations: ["lounge", "stations"] as const,
  session: (sessionId: string) => ["lounge", "session", sessionId] as const,
};

export function useLoungeStations(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: loungeQueryKeys.stations,
    queryFn: ({ signal }) => fetchLoungeStations(signal),
    enabled: options?.enabled ?? true,
    // Keep the map current when another desk checks someone in or out.
    refetchInterval: STATIONS_REFETCH_INTERVAL_MS,
  });
}

export function useLoungeSession(sessionId: string | null, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: loungeQueryKeys.session(sessionId ?? ""),
    queryFn: ({ signal }) => fetchLoungeSession(sessionId!, signal),
    enabled: options?.enabled !== false && !!sessionId,
  });
}

export function useCheckOutLoungeSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (sessionId: string) => checkOutLoungeSession(sessionId),
    // Refresh on failure too: a 409 means someone else already checked them out.
    onSettled: (_data, _error, sessionId) =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: loungeQueryKeys.stations }),
        queryClient.invalidateQueries({ queryKey: loungeQueryKeys.session(sessionId) }),
      ]),
  });
}
