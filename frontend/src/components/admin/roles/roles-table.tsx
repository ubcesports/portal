"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { DataTable, type Column } from "@/components/admin/admin-data-table";
import { EmptyValue } from "@/components/admin/admin-table-cells";
import { RoleBadge } from "@/components/admin/roles/role-badge";
import type { Role } from "@/lib/types/role.types";
import { formatTime } from "@/lib/utils/formatting";

type RolesTableProps = {
  roles: Role[];
  isLoading: boolean;
  isFetching: boolean;
  /* Overrides the message shown when there are no rows. */
  emptyLabel?: string;
};

const COLUMNS: Column<Role>[] = [
  {
    header: "Role",
    cellClassName: "whitespace-nowrap px-4 py-3",
    cell: (role) => (
      <Link href={`/admin/roles/${role.id}`} className="hover:opacity-80">
        <RoleBadge name={role.name} color={role.color} />
      </Link>
    ),
  },
  {
    header: "Description",
    cellClassName: "max-w-md truncate px-4 py-3 text-brand-text-muted",
    cell: (role) => role.description || <EmptyValue />,
  },
  {
    header: "Created",
    cellClassName: "whitespace-nowrap px-4 py-3",
    cell: (role) => (
      <div>
        <p className="text-brand-text-muted">{formatTime(role.created_at)}</p>
        {role.created_by ? (
          <p className="mt-0.5 text-xs text-brand-text-subtle">by {role.created_by.name}</p>
        ) : null}
      </div>
    ),
  },
  {
    header: "Updated",
    cell: (role) => formatTime(role.updated_at),
  },
];

export function RolesTable({ roles, isLoading, isFetching, emptyLabel }: RolesTableProps) {
  const router = useRouter();

  return (
    <DataTable
      data={roles}
      columns={COLUMNS}
      getRowKey={(role) => role.id}
      isLoading={isLoading}
      isFetching={isFetching}
      loadingLabel="Loading roles"
      emptyLabel={emptyLabel ?? "No roles yet. Create one to grant admin permissions to members."}
      onRowClick={(role) => router.push(`/admin/roles/${role.id}`)}
    />
  );
}
