"use client";

import { Loader2, Trash2, X } from "lucide-react";
import { useState } from "react";
import { ActionButton } from "@/components/action-button";
import { StatusBadge } from "@/components/status-badge";
import { formatOptionalTime } from "@/components/admin/admin-table-cells";
import { DataTable, type Column } from "@/components/admin/admin-data-table";
import type { MembershipInvitationView } from "@/components/admin/membership-invitations/membership-invitation-tabs";
import type { MembershipInvitation } from "@/lib/types/admin.types";
import { formatTime } from "@/lib/utils/formatting";

const CAD_FORMAT = new Intl.NumberFormat("en-CA", {
  style: "currency",
  currency: "CAD",
});

type MembershipInvitationsTableProps = {
  invitations: MembershipInvitation[];
  view: MembershipInvitationView;
  isLoading: boolean;
  isFetching: boolean;
  deletingID: string | null;
  onDelete: (id: string) => Promise<void>;
};

export function MembershipInvitationsTable({
  invitations,
  view,
  isLoading,
  isFetching,
  deletingID,
  onDelete,
}: MembershipInvitationsTableProps) {
  const [confirmingID, setConfirmingID] = useState<string | null>(null);

  const columns: Column<MembershipInvitation>[] = [
    {
      header: "Email",
      cellClassName: "whitespace-nowrap px-4 py-3 font-medium text-brand-text",
      cell: (invitation) => invitation.email,
    },
    {
      header: "Membership",
      cellClassName: "whitespace-nowrap px-4 py-3",
      cell: (invitation) => (
        <div>
          <p className="font-medium text-brand-text">{invitation.tier_title}</p>
          <p className="mt-0.5 text-xs uppercase tracking-wide text-brand-text-subtle">
            {invitation.program_name}
          </p>
        </div>
      ),
    },
    {
      header: "Paid",
      cellClassName: "whitespace-nowrap px-4 py-3 font-mono text-brand-text",
      cell: (invitation) => CAD_FORMAT.format(invitation.amount_paid_cents / 100),
    },
    {
      header: "Payment",
      cell: (invitation) => (
        <StatusBadge tone={invitation.payment_method === "cash" ? "success" : "default"}>
          {invitation.payment_method === "cash" ? "Cash" : "E-transfer"}
        </StatusBadge>
      ),
    },
    {
      header: "Invite sent",
      cell: (invitation) => formatOptionalTime(invitation.invitation_sent_at),
    },
    {
      header: "Status",
      cellClassName: "whitespace-nowrap px-4 py-3",
      cell: (invitation) => (
        <div>
          <StatusBadge tone={invitation.done ? "success" : "warning"}>
            {invitation.done ? "Completed" : "Pending"}
          </StatusBadge>
          {invitation.done ? (
            <p className="mt-1 text-xs text-brand-text-subtle">
              {formatTime(invitation.updated_at)}
            </p>
          ) : null}
        </div>
      ),
    },
    {
      header: "Recorded",
      cellClassName: "whitespace-nowrap px-4 py-3",
      cell: (invitation) => (
        <div>
          <p className="text-brand-text-muted">{formatTime(invitation.purchased_at)}</p>
          <p className="mt-0.5 text-xs text-brand-text-subtle">by {invitation.created_by_name}</p>
        </div>
      ),
    },
  ];

  if (view === "pending") {
    columns.push({
      header: "Actions",
      cellClassName: "whitespace-nowrap px-4 py-3 text-right",
      cell: (invitation) => {
        const isConfirming = confirmingID === invitation.id;
        const isDeleting = deletingID === invitation.id;

        if (isConfirming) {
          return (
            <div className="flex items-center justify-end gap-2">
              <ActionButton
                aria-label={`Cancel deleting invitation for ${invitation.email}`}
                onClick={() => setConfirmingID(null)}
                disabled={isDeleting}
                icon={<X aria-hidden="true" className="size-4" />}
                className="h-9 px-3"
              >
                Cancel
              </ActionButton>
              <ActionButton
                onClick={async () => {
                  try {
                    await onDelete(invitation.id);
                    setConfirmingID(null);
                  } catch {
                    // The shared API client displays the server error.
                  }
                }}
                loading={isDeleting}
                icon={<Trash2 aria-hidden="true" className="size-4" />}
                loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
                className="h-9 border-red-400/50 bg-red-400/10 px-3 text-red-100 hover:border-red-300 hover:bg-red-400/20"
              >
                Confirm delete
              </ActionButton>
            </div>
          );
        }

        return (
          <ActionButton
            aria-label={`Delete invitation for ${invitation.email}`}
            onClick={() => setConfirmingID(invitation.id)}
            disabled={deletingID !== null}
            icon={<Trash2 aria-hidden="true" className="size-4" />}
            className="h-9 border-red-400/30 px-3 text-red-200 hover:border-red-300 hover:bg-red-400/10"
          >
            Delete
          </ActionButton>
        );
      },
    });
  }

  return (
    <DataTable
      data={invitations}
      columns={columns}
      getRowKey={(invitation) => invitation.id}
      isLoading={isLoading}
      isFetching={isFetching}
      loadingLabel={`Loading ${view} membership invitations`}
      emptyLabel={
        view === "pending"
          ? "No pending invitations. New in-person purchases will appear here."
          : "No completed invitations yet. Redeemed memberships will appear here."
      }
    />
  );
}
