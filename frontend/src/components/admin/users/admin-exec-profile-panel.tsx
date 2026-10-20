"use client";

import { ChevronRight, Loader2, RefreshCw, Save, Settings2 } from "lucide-react";
import { useState, type FormEvent } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { SurfacePanel } from "@/components/surface-panel";
import { useAdminExecProfile, useUpdateAdminExecProfile } from "@/lib/admin/admin.hook";
import type { AdminExecProfile, UpdateAdminExecProfileRequest } from "@/lib/types/admin.types";
import type { ExecDisplayGroup } from "@/lib/types/exec-profile.types";

const FIELD_CLASS_NAME =
  "h-11 w-full border border-brand-border bg-brand-surface px-3 text-sm text-brand-text outline-none transition focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:cursor-not-allowed disabled:opacity-60";

const DISPLAY_GROUP_OPTIONS: { value: ExecDisplayGroup; label: string }[] = [
  { value: "president", label: "President" },
  { value: "board", label: "Board" },
  { value: "central_director", label: "Central Director" },
  { value: "game_director", label: "Game Director" },
  { value: "executive", label: "Executive" },
];

type ExecDisplaySettingsFormProps = {
  userId: string;
  profile: AdminExecProfile;
};

function ExecDisplaySettingsForm({ userId, profile }: ExecDisplaySettingsFormProps) {
  const [displayGroup, setDisplayGroup] = useState(profile.display_group);
  const [displayOrder, setDisplayOrder] = useState(String(profile.display_order));
  const { mutateAsync: updateProfile, isPending } = useUpdateAdminExecProfile(userId);

  const parsedOrder = Number(displayOrder);
  const validOrder =
    displayOrder.trim() !== "" && Number.isInteger(parsedOrder) && parsedOrder >= 0;
  const isDirty =
    displayGroup !== profile.display_group || (validOrder && parsedOrder !== profile.display_order);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!validOrder) {
      toast.error("Display order must be a whole number of zero or greater");
      return;
    }

    const body: UpdateAdminExecProfileRequest = {};
    if (displayGroup !== profile.display_group) body.display_group = displayGroup;
    if (parsedOrder !== profile.display_order) body.display_order = parsedOrder;

    if (Object.keys(body).length === 0) return;

    try {
      await updateProfile(body);
      toast.success("Executive display settings updated");
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <form onSubmit={handleSubmit} className="grid gap-5 p-5 sm:p-6">
      <div className="flex flex-col gap-1 border-l-2 border-brand-primary bg-white/3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
        <span className="text-xs font-semibold uppercase tracking-wide text-brand-text-subtle">
          Exec-managed title
        </span>
        <span className="text-sm font-medium text-brand-text">{profile.title}</span>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(12rem,0.45fr)_auto] lg:items-end">
        <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
          <span>Display group</span>
          <select
            value={displayGroup}
            onChange={(event) => setDisplayGroup(event.target.value as ExecDisplayGroup)}
            disabled={isPending}
            className={FIELD_CLASS_NAME}
          >
            {DISPLAY_GROUP_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
          <span>Display order</span>
          <input
            type="number"
            min={0}
            step={1}
            inputMode="numeric"
            value={displayOrder}
            onChange={(event) => setDisplayOrder(event.target.value)}
            disabled={isPending}
            className={FIELD_CLASS_NAME}
          />
        </label>

        <ActionButton
          type="submit"
          disabled={!isDirty || !validOrder}
          loading={isPending}
          icon={<Save aria-hidden="true" className="size-4" />}
          loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
          className="h-11 border-brand-primary bg-brand-primary hover:border-brand-primary-hover hover:bg-brand-primary-hover"
        >
          Save settings
        </ActionButton>
      </div>

      <p className="text-xs text-brand-text-subtle">
        Lower order numbers appear first within the selected display group.
      </p>
    </form>
  );
}

export function AdminExecProfilePanel({ userId }: { userId: string }) {
  const { data: profile, isPending, isError, isFetching, refetch } = useAdminExecProfile(userId);

  return (
    <SurfacePanel className="overflow-hidden bg-transparent">
      <details open className="group/admin-exec-profile">
        <summary className="flex cursor-pointer list-none items-center gap-3 bg-amber-300/8 px-5 py-4 transition hover:bg-amber-300/12 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-brand-primary group-open/admin-exec-profile:border-b group-open/admin-exec-profile:border-brand-border [&::-webkit-details-marker]:hidden">
          <ChevronRight
            aria-hidden="true"
            className="size-4 shrink-0 text-amber-200 transition-transform group-open/admin-exec-profile:rotate-90"
          />
          <Settings2 aria-hidden="true" className="size-4 shrink-0 text-amber-200" />
          <span>
            <span className="block text-base font-semibold text-brand-text">
              Executive display settings
            </span>
            <span className="mt-1 block text-sm text-brand-text-subtle">
              Choose where this executive appears on the public roster.
            </span>
          </span>
        </summary>

        {isPending ? (
          <div className="flex items-center gap-3 px-5 py-8 text-sm text-brand-text-muted">
            <Loader2 aria-hidden="true" className="size-4 animate-spin" />
            Loading executive display settings
          </div>
        ) : isError || !profile ? (
          <div
            role="alert"
            className="flex flex-col gap-3 px-5 py-6 sm:flex-row sm:items-center sm:justify-between"
          >
            <p className="text-sm text-brand-text-muted">
              Executive display settings could not be loaded.
            </p>
            <ActionButton
              onClick={() => void refetch()}
              loading={isFetching}
              icon={<RefreshCw aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            >
              Try again
            </ActionButton>
          </div>
        ) : (
          <ExecDisplaySettingsForm
            key={`${profile.display_group}:${profile.display_order}:${profile.title}`}
            userId={userId}
            profile={profile}
          />
        )}
      </details>
    </SurfacePanel>
  );
}
