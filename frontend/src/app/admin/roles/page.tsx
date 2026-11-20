"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { RolesTable } from "@/components/admin/roles/roles-table";
import { useRoles } from "@/lib/admin/admin.roles.hook";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import { AdminTablePage } from "../admin-table-page";

export default function AdminRolesPage() {
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const rolesQuery = useRoles({ enabled: isAdmin });
  const roles = rolesQuery.data ?? [];

  return (
    <AdminTablePage
      title="Roles"
      description="Group admin permissions into roles and give them to members."
      isLoading={isProfilePending || (isAdmin && rolesQuery.isPending)}
      loadingLabel="Loading roles"
      toolbar={
        <div className="flex justify-end border-b border-brand-border px-5 py-4 sm:px-6">
          <Link
            href="/admin/roles/new"
            className="inline-flex h-10 items-center gap-2 border border-brand-primary bg-brand-primary px-4 text-sm font-semibold text-white transition hover:border-brand-primary-hover hover:bg-brand-primary-hover"
          >
            <Plus aria-hidden="true" className="size-4" />
            New role
          </Link>
        </div>
      }
      table={
        <RolesTable
          roles={roles}
          isLoading={rolesQuery.isPending}
          isFetching={rolesQuery.isFetching}
        />
      }
      pagination={
        <div className="flex min-h-14 items-center border-t border-brand-border px-5 text-sm text-brand-text-subtle sm:px-6">
          <span className="font-mono text-brand-text">{roles.length}</span>
          <span className="ml-2">{roles.length === 1 ? "role" : "roles"}</span>
        </div>
      }
    />
  );
}
