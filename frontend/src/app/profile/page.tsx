"use client";

import { ExternalLink, Loader2, LogOut, RefreshCw } from "lucide-react";
import { useMutation } from "@tanstack/react-query";
import { ActionButton } from "@/components/action-button";
import { ActionLink } from "@/components/action-link";
import { DetailRow } from "@/components/detail-row";
import { BasePage } from "@/components/layout/base-page";
import { MembershipHistoryItem } from "@/components/membership/membership-history-item";
import { MembershipLoadError } from "@/components/membership/membership-load-error";
import { ExecProfilePanel } from "@/components/profile/exec-profile-panel";
import { StatusBadge } from "@/components/status-badge";
import { SummaryTile } from "@/components/summary-tile";
import { SurfacePanel } from "@/components/surface-panel";
import { redirectToSignIn } from "@/lib/auth";
import { useSignOut } from "@/lib/use-sign-out.hook";
import { formatDate, getInitials } from "@/lib/utils/formatting";
import { getGroupBadgeClass, titleCase } from "@/lib/utils/groups";
import { useProfile } from "@/lib/profile.hook";
import { useAllMemberships } from "@/lib/membership.hook";
import type { Membership } from "@/lib/types/membership.types";
import Image from "next/image";

const ZETROVA_ACCOUNT_URL =
  process.env.NEXT_PUBLIC_ZETROVA_ACCOUNT_URL || "https://id.zetrova.com/dashboard";

function isActiveMembership(membership: Membership) {
  return !membership.cancelled_at && new Date(membership.expires_at) > new Date();
}

type MembershipSectionProps = {
  title: string;
  description: string;
  memberships: Membership[];
  emptyMessage: string;
};

function MembershipSection({
  title,
  description,
  memberships,
  emptyMessage,
}: MembershipSectionProps) {
  return (
    <SurfacePanel>
      <div className="px-5 py-4">
        <h3 className="text-base font-semibold text-brand-text">{title}</h3>
        <p className="mt-1 text-sm text-brand-text-subtle">{description}</p>
      </div>

      {memberships.length === 0 ? (
        <p className="border-t border-brand-border px-5 py-6 text-sm text-brand-text-muted">
          {emptyMessage}
        </p>
      ) : (
        <ul className="divide-y divide-brand-border border-t border-brand-border">
          {memberships.map((membership) => (
            <MembershipHistoryItem key={membership.id} membership={membership} />
          ))}
        </ul>
      )}
    </SurfacePanel>
  );
}

export default function ProfilePage() {
  const { data: profile, isPending } = useProfile();
  const {
    data: memberships,
    isPending: membershipsPending,
    isError: membershipsError,
    isFetching: membershipsFetching,
    refetch: refetchMemberships,
  } = useAllMemberships();

  const { mutate: signOut, error: signOutError, isPending: signOutPending } = useSignOut();

  const {
    mutate: syncAccount,
    error: syncAccountError,
    isPending: syncAccountPending,
  } = useMutation({
    mutationFn: async () => await redirectToSignIn(window.location.href),
  });

  const error = signOutError
    ? "Sign out failed. Try again."
    : syncAccountError
      ? "Unable to sync account. Try again."
      : null;

  const displayName = profile?.name ?? profile?.email ?? "Profile";
  const activeMemberships = memberships?.filter(isActiveMembership) ?? [];
  const membershipHistory =
    memberships?.filter((membership) => !isActiveMembership(membership)) ?? [];
  const isExecutive =
    profile?.groups.some((group) => ["executive", "director", "board"].includes(group)) ?? false;

  const studentBadge = profile?.isStudent ? (
    <StatusBadge tone="success">Student</StatusBadge>
  ) : (
    <StatusBadge tone="muted">Non-student</StatusBadge>
  );

  return (
    <BasePage>
      <div className="flex flex-1 items-center py-12">
        <section className="mx-auto w-full max-w-6xl">
          <div className="mt-10 border border-brand-border bg-brand-surface/80 shadow-2xl shadow-black/25">
            <div className="flex flex-col gap-4 border-b border-brand-border px-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
              <div>
                <h2 className="text-lg font-semibold text-brand-text">UBCEA Account</h2>
                <p className="mt-1 text-sm text-brand-text-subtle">
                  Your profile and account status.
                </p>
              </div>
              {profile ? (
                <ActionButton
                  onClick={() => signOut()}
                  loading={signOutPending}
                  icon={<LogOut aria-hidden="true" className="size-4" />}
                  loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
                >
                  {signOutPending ? "Signing out" : "Sign out"}
                </ActionButton>
              ) : null}
            </div>

            {isPending ? (
              <div className="flex min-h-56 items-center justify-center gap-3 px-6 py-12 text-brand-text-muted">
                <Loader2 aria-hidden="true" className="size-5 animate-spin" />
                <span>Loading profile</span>
              </div>
            ) : error ? (
              <div className="px-6 py-12 text-brand-text-muted">{error}</div>
            ) : profile ? (
              <div className="p-5 sm:p-6">
                <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(360px,0.78fr)]">
                  <SurfacePanel className="p-5 sm:p-6">
                    <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
                      {profile.avatarUrl ? (
                        <Image
                          src={profile.avatarUrl}
                          alt=""
                          width={72}
                          height={72}
                          className="size-18 shrink-0 border border-brand-primary/40 bg-brand-primary/15 object-cover"
                        />
                      ) : (
                        <div className="flex size-18 shrink-0 items-center justify-center border border-brand-primary/40 bg-brand-primary/15 text-2xl font-semibold text-brand-text">
                          {getInitials(profile.name, profile.email)}
                        </div>
                      )}
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-3">
                          <h3 className="wrap-break-word text-2xl font-semibold text-brand-text">
                            {displayName}
                          </h3>
                          <StatusBadge tone={profile.role === "admin" ? "warning" : "default"}>
                            {titleCase(profile.role)}
                          </StatusBadge>
                        </div>
                        <p className="mt-2 wrap-break-word text-sm text-brand-text-muted">
                          {profile.createdAt
                            ? `Member since ${formatDate(profile.createdAt)}`
                            : "Membership start date unavailable"}
                        </p>
                      </div>
                    </div>

                    <div className="mt-6 grid gap-3.5">
                      <SummaryTile
                        label="Assigned groups"
                        detail={
                          profile.groups.length > 0 ? (
                            <div className="flex flex-wrap gap-2">
                              {profile.groups.map((group) => (
                                <StatusBadge
                                  key={group}
                                  tone="default"
                                  className={getGroupBadgeClass(group)}
                                >
                                  {titleCase(group)}
                                </StatusBadge>
                              ))}
                            </div>
                          ) : (
                            "No groups assigned"
                          )
                        }
                        tone="default"
                      />
                    </div>
                  </SurfacePanel>

                  <SurfacePanel className="flex flex-col">
                    <div className="px-5 py-4">
                      <h3 className="text-base font-semibold text-brand-text">Account Details</h3>
                    </div>
                    <dl>
                      <DetailRow label="Email">
                        <span className="wrap-break-word">{profile.email}</span>
                      </DetailRow>
                      <DetailRow label="Student status">{studentBadge}</DetailRow>
                      <DetailRow label="Student ID">
                        {profile.studentId ? (
                          <span className="wrap-break-word font-mono">{profile.studentId}</span>
                        ) : (
                          <StatusBadge tone="muted">Not provided</StatusBadge>
                        )}
                      </DetailRow>
                    </dl>
                    <div className="mt-auto grid gap-3 border-t border-brand-border px-5 py-4 sm:grid-cols-2">
                      <ActionLink
                        href={ZETROVA_ACCOUNT_URL}
                        icon={<ExternalLink aria-hidden="true" className="size-4" />}
                      >
                        Manage
                      </ActionLink>
                      <ActionButton
                        onClick={() => syncAccount()}
                        loading={syncAccountPending}
                        icon={<RefreshCw aria-hidden="true" className="size-4" />}
                        loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
                      >
                        {syncAccountPending ? "Syncing" : "Sync"}
                      </ActionButton>
                    </div>
                  </SurfacePanel>
                </div>

                {isExecutive ? <ExecProfilePanel /> : null}

                <div id="membership" className="mt-6 scroll-mt-28">
                  {membershipsPending ? (
                    <SurfacePanel>
                      <div className="px-5 py-4">
                        <h3 className="text-base font-semibold text-brand-text">Memberships</h3>
                      </div>
                      <div className="flex items-center gap-3 border-t border-brand-border px-5 py-6 text-brand-text-muted">
                        <Loader2 aria-hidden="true" className="size-4 animate-spin" />
                        <span className="text-sm">Loading memberships</span>
                      </div>
                    </SurfacePanel>
                  ) : membershipsError ? (
                    <MembershipLoadError
                      isRetrying={membershipsFetching}
                      onRetry={() => void refetchMemberships()}
                    />
                  ) : (
                    <div className="grid gap-6">
                      <MembershipSection
                        title="Active memberships"
                        description="Memberships that are currently available to your account."
                        memberships={activeMemberships}
                        emptyMessage="No active memberships."
                      />
                      <MembershipSection
                        title="Membership history"
                        description="Your expired and cancelled memberships."
                        memberships={membershipHistory}
                        emptyMessage="No expired or cancelled memberships."
                      />
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="px-6 py-12 text-brand-text-muted">
                No profile details were returned.
              </div>
            )}
          </div>
        </section>
      </div>
    </BasePage>
  );
}
