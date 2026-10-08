"use client";

import { toast } from "sonner";
import { useState } from "react";
import { MembershipInvitationForm } from "@/components/admin/membership-invitations/membership-invitation-form";
import {
  MembershipInvitationTabs,
  type MembershipInvitationView,
} from "@/components/admin/membership-invitations/membership-invitation-tabs";
import { MembershipInvitationsTable } from "@/components/admin/membership-invitations/membership-invitations-table";
import {
  useAdminMembershipTierOptions,
  useCreateMembershipInvitation,
  useDeleteMembershipInvitation,
  useMembershipInvitations,
} from "@/lib/admin/admin.hook";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import type { CreateMembershipInvitationRequest } from "@/lib/types/admin.types";
import { AdminTablePage } from "../admin-table-page";

export default function MembershipInvitationsPage() {
  const [activeView, setActiveView] = useState<MembershipInvitationView>("pending");
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const invitationsQuery = useMembershipInvitations({ enabled: isAdmin });
  const tiersQuery = useAdminMembershipTierOptions({ enabled: isAdmin });
  const createInvitation = useCreateMembershipInvitation();
  const deleteInvitation = useDeleteMembershipInvitation();

  const allInvitations = invitationsQuery.data ?? [];
  const pendingInvitations = allInvitations.filter((invitation) => !invitation.done);
  const completedInvitations = allInvitations.filter((invitation) => invitation.done);
  const invitations = activeView === "pending" ? pendingInvitations : completedInvitations;

  const handleCreate = async (request: CreateMembershipInvitationRequest) => {
    await createInvitation.mutateAsync(request);
    toast.success("Membership invitation created", {
      description: `An invitation was sent to ${request.email}.`,
    });
  };

  const handleDelete = async (id: string) => {
    await deleteInvitation.mutateAsync(id);
    toast.success("Membership invitation deleted");
  };

  return (
    <AdminTablePage
      title="Membership Invitations"
      description="Record in-person purchases and track invitations awaiting account signup."
      isLoading={
        isProfilePending ||
        (isAdmin && invitationsQuery.isPending && !invitationsQuery.isPlaceholderData)
      }
      loadingLabel="Loading membership invitations"
      toolbar={
        <>
          <MembershipInvitationTabs
            activeView={activeView}
            pendingCount={pendingInvitations.length}
            completedCount={completedInvitations.length}
            onChange={setActiveView}
          />
          {activeView === "pending" ? (
            <MembershipInvitationForm
              tierOptions={tiersQuery.data ?? []}
              isLoadingTiers={tiersQuery.isPending}
              isSubmitting={createInvitation.isPending}
              onSubmit={handleCreate}
            />
          ) : null}
        </>
      }
      table={
        <div
          id="membership-invitations-panel"
          role="tabpanel"
          className="flex min-h-0 w-full flex-1"
        >
          <MembershipInvitationsTable
            invitations={invitations}
            view={activeView}
            isLoading={invitationsQuery.isPending}
            isFetching={invitationsQuery.isFetching}
            deletingID={deleteInvitation.isPending ? (deleteInvitation.variables ?? null) : null}
            onDelete={handleDelete}
          />
        </div>
      }
      pagination={
        <div className="flex min-h-14 items-center border-t border-brand-border px-5 text-sm text-brand-text-subtle sm:px-6">
          <span className="font-mono text-brand-text">{invitations.length}</span>
          <span className="ml-2">
            {activeView === "pending"
              ? invitations.length === 1
                ? "invitation waiting to be redeemed"
                : "invitations waiting to be redeemed"
              : invitations.length === 1
                ? "completed invitation"
                : "completed invitations"}
          </span>
        </div>
      }
    />
  );
}
