"use client";

import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useEffect, type ReactNode } from "react";
import { isAxiosError } from "axios";
import { UserMembershipsPanel } from "@/components/admin/users/user-memberships-panel";
import { UserProfilePanel } from "@/components/admin/users/user-profile-panel";
import { AdminExecProfilePanel } from "@/components/admin/users/admin-exec-profile-panel";
import { BasePage } from "@/components/layout/base-page";
import { MembershipLoadError } from "@/components/membership/membership-load-error";
import { useUpdateUser, useUser, useUserMemberships } from "@/lib/admin/admin.hook";
import type { UpdateUserRequest } from "@/lib/types/admin.types";
import { useProfile } from "@/lib/profile.hook";

type AdminUserDetailPageProps = {
  params: Promise<{ userId: string }>;
};

function CenteredPanel({ children }: { children: ReactNode }) {
  return (
    <BasePage>
      <div className="flex flex-1 items-center py-6">
        <section className="mx-auto flex min-h-[60vh] w-full items-center justify-center border border-brand-border bg-brand-surface/80 shadow-2xl shadow-black/25">
          {children}
        </section>
      </div>
    </BasePage>
  );
}

export default function AdminUserDetailPage({ params }: AdminUserDetailPageProps) {
  const { userId } = use(params);
  const router = useRouter();
  const { data: profile, isPending: isProfilePending } = useProfile();

  const isAdmin = profile?.role === "admin";

  useEffect(() => {
    if (!isProfilePending && profile && profile.role !== "admin") {
      router.replace("/403");
    }
  }, [isProfilePending, profile, router]);

  const {
    data: user,
    isPending: isUserPending,
    error: userError,
  } = useUser(userId, { enabled: isAdmin });
  const {
    data: memberships,
    isPending: areMembershipsPending,
    isError: membershipsError,
    isFetching: membershipsFetching,
    refetch: refetchMemberships,
  } = useUserMemberships(userId, { enabled: isAdmin });
  const { mutateAsync: updateUser, isPending: isSaving } = useUpdateUser(userId);

  const handleSave = (body: UpdateUserRequest) => updateUser(body);

  if (isProfilePending || (isAdmin && isUserPending && !userError)) {
    return (
      <CenteredPanel>
        <div className="flex items-center gap-3 text-brand-text-muted">
          <Loader2 aria-hidden="true" className="size-5 animate-spin" />
          <span>Loading user</span>
        </div>
      </CenteredPanel>
    );
  }

  if (!isAdmin) {
    return null;
  }

  if (userError) {
    const notFound = isAxiosError(userError) && userError.response?.status === 404;

    return (
      <CenteredPanel>
        <div className="flex flex-col items-center gap-3 px-6 py-12 text-center">
          <p className="text-sm text-brand-text-muted">
            {notFound ? "This user does not exist." : "Unable to load this user."}
          </p>
          <Link
            href="/admin/users"
            className="inline-flex h-10 items-center gap-2 border border-brand-border px-4 text-sm font-semibold text-brand-text transition hover:border-brand-text-muted hover:bg-white/5"
          >
            <ArrowLeft aria-hidden="true" className="size-4" />
            Back to users
          </Link>
        </div>
      </CenteredPanel>
    );
  }

  if (!user) {
    return null;
  }

  const hasExecGroup = user.groups.some((group) =>
    ["executive", "director", "board"].includes(group),
  );

  return (
    <BasePage>
      <div className="flex flex-1 flex-col py-6">
        <section className="mx-auto flex w-full max-w-5xl flex-col gap-4">
          <Link
            href="/admin/users"
            className="inline-flex w-fit items-center gap-2 text-sm text-brand-text-subtle transition hover:text-brand-text"
          >
            <ArrowLeft aria-hidden="true" className="size-4" />
            Back to users
          </Link>

          <UserProfilePanel user={user} onSave={handleSave} isSaving={isSaving} />

          {hasExecGroup ? <AdminExecProfilePanel userId={userId} /> : null}

          {areMembershipsPending ? (
            <div className="flex items-center gap-3 border border-brand-border px-5 py-6 text-sm text-brand-text-muted">
              <Loader2 aria-hidden="true" className="size-4 animate-spin" />
              <span>Loading memberships</span>
            </div>
          ) : membershipsError ? (
            <MembershipLoadError
              isRetrying={membershipsFetching}
              onRetry={() => void refetchMemberships()}
            />
          ) : (
            <UserMembershipsPanel
              userId={userId}
              memberships={memberships ?? []}
              onSave={handleSave}
              isSaving={isSaving}
            />
          )}
        </section>
      </div>
    </BasePage>
  );
}
