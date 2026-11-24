"use client";

import { Loader2 } from "lucide-react";
import { useState } from "react";
import { BasePage } from "@/components/layout/base-page";
import { LoungeMapSection } from "@/components/lounge/lounge-map-section";
import { useRequireAdmin } from "@/lib/admin/require.admin";

// Admin-only for now; switch to the lounge permissions once the new roles system lands.
export default function LoungePage() {
  const { isAdmin } = useRequireAdmin();
  // Shared so other lounge sections (e.g. check-in) can use the station picked on the map.
  const [selectedStationId, setSelectedStationId] = useState<string | null>(null);

  return (
    <BasePage>
      <div className="flex flex-col gap-6 py-6">
        <div>
          <h1 className="text-2xl font-semibold text-brand-text">Lounge</h1>
          <p className="mt-1 text-sm text-brand-text-subtle">
            See which stations are free and manage active sessions.
          </p>
        </div>

        {isAdmin ? (
          <div className="flex flex-col gap-6">
            <LoungeMapSection
              selectedStationId={selectedStationId}
              onSelectStation={setSelectedStationId}
            />
          </div>
        ) : (
          <div className="flex items-center gap-3 text-brand-text-muted">
            <Loader2 aria-hidden="true" className="size-5 animate-spin" />
            <span>Loading lounge</span>
          </div>
        )}
      </div>
    </BasePage>
  );
}
