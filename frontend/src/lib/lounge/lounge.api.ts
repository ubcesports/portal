import apiClient from "@/lib/client";
import type { LoungeSession, LoungeSessionDetails, LoungeStation } from "@/lib/types/lounge.types";

export async function fetchLoungeStations(signal?: AbortSignal) {
  const response = await apiClient.get<LoungeStation[]>("/lounge/stations", { signal });
  return response.data;
}

export async function fetchLoungeSession(sessionId: string, signal?: AbortSignal) {
  const response = await apiClient.get<LoungeSessionDetails>(`/lounge/sessions/${sessionId}`, {
    signal,
  });
  return response.data;
}

export async function checkOutLoungeSession(sessionId: string) {
  const response = await apiClient.post<LoungeSession>(`/lounge/sessions/${sessionId}/check-out`);
  return response.data;
}
