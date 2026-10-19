"use client";

import { CheckCircle2, Clock3 } from "lucide-react";

export type MembershipInvitationView = "pending" | "completed";

type MembershipInvitationTabsProps = {
  activeView: MembershipInvitationView;
  pendingCount: number;
  completedCount: number;
  onChange: (view: MembershipInvitationView) => void;
};

const TABS = [
  { id: "pending", label: "Pending", icon: Clock3 },
  { id: "completed", label: "Completed", icon: CheckCircle2 },
] as const;

export function MembershipInvitationTabs({
  activeView,
  pendingCount,
  completedCount,
  onChange,
}: MembershipInvitationTabsProps) {
  return (
    <div
      role="tablist"
      aria-label="Membership invitation status"
      className="flex gap-1 border-b border-brand-border px-5 pt-3 sm:px-6"
    >
      {TABS.map((tab) => {
        const Icon = tab.icon;
        const isActive = activeView === tab.id;
        const count = tab.id === "pending" ? pendingCount : completedCount;

        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            aria-controls="membership-invitations-panel"
            onClick={() => onChange(tab.id)}
            className={`relative flex min-h-11 items-center gap-2 px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary/60 ${
              isActive
                ? "text-brand-text"
                : "text-brand-text-subtle hover:bg-white/[0.035] hover:text-brand-text"
            }`}
          >
            <Icon aria-hidden="true" className="size-4" />
            <span>{tab.label}</span>
            <span
              className={`min-w-6 border px-1.5 py-0.5 font-mono text-[11px] leading-none ${
                isActive
                  ? "border-brand-primary/50 bg-brand-primary/15 text-brand-text"
                  : "border-brand-border bg-white/3 text-brand-text-subtle"
              }`}
            >
              {count}
            </span>
            {isActive ? (
              <span
                aria-hidden="true"
                className="absolute inset-x-0 -bottom-px h-0.5 bg-brand-primary"
              />
            ) : null}
          </button>
        );
      })}
    </div>
  );
}
