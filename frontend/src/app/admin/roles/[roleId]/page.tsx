"use client";

import { ArrowLeft, Loader2, Trash2, X } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useState, type ReactNode } from "react";
import { isAxiosError } from "axios";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { RoleForm } from "@/components/admin/roles/role-form";
import { RoleUsersPanel } from "@/components/admin/roles/role-users-panel";
import { BasePage } from "@/components/layout/base-page";
import { useDeleteRole, useRole, useUpdateRole } from "@/lib/admin/admin.roles.hook";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import type { RoleDetails, UpdateRoleRequest } from "@/lib/types/role.types";

type AdminRoleDetailPageProps = {
  params: Promise<{ roleId: string }>;
};

function PageShell({ children }: { children: ReactNode }) {
  return (
    <BasePage>
      <div className="flex flex-1 flex-col py-6">
        <section className="mx-auto flex w-full max-w-5xl flex-col gap-4">
          <Link
            href="/admin/roles"
            className="inline-flex w-fit items-center gap-2 text-sm text-brand-text-subtle transition hover:text-brand-text"
          >
            <ArrowLeft aria-hidden="true" className="size-4" />
            Back to roles
          </Link>
          {children}
        </section>
      </div>
    </BasePage>
  );
}

function DeleteRoleButton({ role, disabled }: { role: RoleDetails; disabled?: boolean }) {
  const router = useRouter();
  const [isConfirming, setIsConfirming] = useState(false);
  const deleteRole = useDeleteRole();

  const handleDelete = async () => {
    try {
      await deleteRole.mutateAsync(role.id);
    } catch {
      // The shared API client displays the server error.
      return;
    }
    toast.success(`Deleted ${role.name}`);
    router.push("/admin/roles");
  };

  if (!isConfirming) {
    return (
      <ActionButton
        onClick={() => setIsConfirming(true)}
        disabled={disabled}
        icon={<Trash2 aria-hidden="true" className="size-4" />}
        className="h-11 border-red-400/30 px-4 text-red-200 hover:border-red-300 hover:bg-red-400/10"
      >
        Delete role
      </ActionButton>
    );
  }

  const userCount = role.users_with_role.length;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-brand-text-muted">
        {userCount > 0
          ? `This removes it from ${userCount} ${userCount === 1 ? "user" : "users"}.`
          : "This cannot be undone."}
      </span>
      <ActionButton
        onClick={() => setIsConfirming(false)}
        disabled={deleteRole.isPending}
        icon={<X aria-hidden="true" className="size-4" />}
        className="h-11 px-3"
      >
        Cancel
      </ActionButton>
      <ActionButton
        onClick={handleDelete}
        disabled={disabled}
        loading={deleteRole.isPending}
        icon={<Trash2 aria-hidden="true" className="size-4" />}
        loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
        className="h-11 border-red-400/50 bg-red-400/10 px-3 text-red-100 hover:border-red-300 hover:bg-red-400/20"
      >
        Confirm delete
      </ActionButton>
    </div>
  );
}

export default function AdminRoleDetailPage({ params }: AdminRoleDetailPageProps) {
  const { roleId } = use(params);
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const { data: role, isPending: isRolePending, error } = useRole(roleId, { enabled: isAdmin });
  const updateRole = useUpdateRole(roleId);

  const handleUpdate = async (body: UpdateRoleRequest) => {
    const updated = await updateRole.mutateAsync(body);
    toast.success(`Saved ${updated.name}`);
  };

  if (isProfilePending || (isAdmin && isRolePending && !error)) {
    return (
      <PageShell>
        <div className="flex items-center gap-3 text-brand-text-muted">
          <Loader2 aria-hidden="true" className="size-5 animate-spin" />
          <span>Loading role</span>
        </div>
      </PageShell>
    );
  }

  if (!isAdmin) {
    return null;
  }

  if (error || !role) {
    const notFound = isAxiosError(error) && error.response?.status === 404;

    return (
      <PageShell>
        <p className="border border-brand-border px-5 py-12 text-center text-sm text-brand-text-muted">
          {notFound ? "This role does not exist." : "Unable to load this role."}
        </p>
      </PageShell>
    );
  }

  return (
    <PageShell>
      <RoleForm
        // Reset the draft whenever the saved role changes.
        key={role.updated_at}
        role={role}
        title={role.name}
        description="Edit the details and permissions. Changes apply to everyone with this role."
        submitLabel={updateRole.isPending ? "Saving" : "Save changes"}
        isSubmitting={updateRole.isPending}
        onSubmit={handleUpdate}
        footer={<DeleteRoleButton role={role} disabled={updateRole.isPending} />}
      />
      <RoleUsersPanel role={role} />
    </PageShell>
  );
}
