"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { RolesTable } from "@/components/admin/roles/roles-table";
import { SearchField } from "@/components/toolbar/search-field";
import { ToolbarContainer } from "@/components/toolbar/toolbar-container";
import { ToolbarRow } from "@/components/toolbar/toolbar-row";
import { useRoles } from "@/lib/admin/admin.roles.hook";
import { useRequireAdmin } from "@/lib/admin/require.admin";
import { AdminTablePage } from "../admin-table-page";

export default function AdminRolesPage() {
  const [search, setSearch] = useState("");
  const { isAdmin, isProfilePending } = useRequireAdmin();
  const rolesQuery = useRoles({ enabled: isAdmin });
  const roles = rolesQuery.data ?? [];

  // Roles are few and already loaded, so filter on the client.
  const query = search.trim().toLowerCase();
  const filteredRoles = query
    ? roles.filter(
        (role) =>
          role.name.toLowerCase().includes(query) ||
          role.description?.toLowerCase().includes(query),
      )
    : roles;

  return (
    <AdminTablePage
      title="Roles"
      description="Group admin permissions into roles and give them to members."
      isLoading={isProfilePending || (isAdmin && rolesQuery.isPending)}
      loadingLabel="Loading roles"
      toolbar={
        <ToolbarContainer>
          <ToolbarRow justify>
            <SearchField
              type="search"
              label="Search roles"
              placeholder="Search by name or description"
              value={search}
              onChange={setSearch}
              className="w-full max-w-md"
            />
            <Link
              href="/admin/roles/new"
              className="inline-flex h-10 w-fit items-center gap-2 border border-brand-primary bg-brand-primary px-4 text-sm font-semibold text-white transition hover:border-brand-primary-hover hover:bg-brand-primary-hover"
            >
              <Plus aria-hidden="true" className="size-4" />
              New role
            </Link>
          </ToolbarRow>
        </ToolbarContainer>
      }
      table={
        <RolesTable
          roles={filteredRoles}
          isLoading={rolesQuery.isPending}
          isFetching={rolesQuery.isFetching}
          emptyLabel={query && roles.length > 0 ? "No roles match your search." : undefined}
        />
      }
      pagination={
        <div className="flex min-h-14 items-center border-t border-brand-border px-5 text-sm text-brand-text-subtle sm:px-6">
          {query ? (
            <>
              <span className="font-mono text-brand-text">{filteredRoles.length}</span>
              <span className="ml-2">of</span>
              <span className="ml-2 font-mono text-brand-text">{roles.length}</span>
              <span className="ml-2">{roles.length === 1 ? "role" : "roles"}</span>
            </>
          ) : (
            <>
              <span className="font-mono text-brand-text">{roles.length}</span>
              <span className="ml-2">{roles.length === 1 ? "role" : "roles"}</span>
            </>
          )}
        </div>
      }
    />
  );
}
