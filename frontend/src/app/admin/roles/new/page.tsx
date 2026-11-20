"use client";

import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { RoleForm } from "@/components/admin/roles/role-form";
import { BasePage } from "@/components/layout/base-page";
import { useCreateRole } from "@/lib/admin/admin.roles.hook";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import type { CreateRoleRequest } from "@/lib/types/role.types";

export default function NewRolePage() {
  const router = useRouter();
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const createRole = useCreateRole();

  const handleCreate = async (body: CreateRoleRequest) => {
    const role = await createRole.mutateAsync(body);
    toast.success(`Created ${role.name}`);
    router.push(`/admin/roles/${role.id}`);
  };

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

          {isProfilePending ? (
            <div className="flex items-center gap-3 text-brand-text-muted">
              <Loader2 aria-hidden="true" className="size-5 animate-spin" />
              <span>Loading</span>
            </div>
          ) : isAdmin ? (
            <RoleForm
              title="New role"
              description="Name the role and choose what it lets members do. You can add users after it is created."
              submitLabel={createRole.isPending ? "Creating" : "Create role"}
              isSubmitting={createRole.isPending}
              onSubmit={handleCreate}
            />
          ) : null}
        </section>
      </div>
    </BasePage>
  );
}
