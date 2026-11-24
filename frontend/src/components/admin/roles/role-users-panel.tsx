"use client";

import { Loader2, Search, UserMinus, UserPlus, X } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { SurfacePanel } from "@/components/surface-panel";
import { useUsers } from "@/lib/admin/admin.hook";
import { useAssignRole, useUnassignRole } from "@/lib/admin/admin.roles.hook";
import type { RoleDetails } from "@/lib/types/role.types";
import { useDebouncedValue } from "@/lib/use-debounced-value.hook";
import { getInitials } from "@/lib/utils/formatting";

const SEARCH_RESULT_LIMIT = 8;
const MIN_SEARCH_LENGTH = 2;
const NO_FILTERS = {};

function Avatar({ src, name, email }: { src: string | null; name: string; email?: string }) {
  if (src) {
    return (
      <Image
        src={src}
        alt=""
        width={32}
        height={32}
        unoptimized
        className="size-8 shrink-0 rounded-full border border-brand-border object-cover"
      />
    );
  }

  return (
    <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-brand-primary/20 text-xs font-bold text-brand-text">
      {getInitials(name, email ?? "")}
    </span>
  );
}

type RoleUserSearchProps = {
  role: RoleDetails;
};

function RoleUserSearch({ role }: RoleUserSearchProps) {
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search.trim(), 300);
  const assignRole = useAssignRole();
  const canSearch = debouncedSearch.length >= MIN_SEARCH_LENGTH;

  const usersQuery = useUsers(
    canSearch ? { mode: "full_name", value: debouncedSearch } : null,
    NO_FILTERS,
    { limit: SEARCH_RESULT_LIMIT, offset: 0 },
    { enabled: canSearch },
  );

  const assignedIDs = new Set(role.users_with_role.map((user) => user.user_id));
  const results = canSearch ? (usersQuery.data?.users ?? []) : [];

  const handleAssign = async (userId: string, name: string) => {
    try {
      await assignRole.mutateAsync({ roleId: role.id, userId });
      toast.success(`Gave ${role.name} to ${name}`);
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <div className="border-b border-brand-border px-5 py-4 sm:px-6">
      <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
        <span>Add a user</span>
        <div className="relative">
          <Search
            aria-hidden="true"
            className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-brand-text-subtle"
          />
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search by full name"
            className="h-10 w-full border border-brand-border bg-brand-surface pl-9 pr-3 text-sm text-brand-text outline-none transition placeholder:text-brand-text-subtle focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30"
          />
        </div>
      </label>

      {canSearch ? (
        <div className="mt-3 border border-brand-border">
          {usersQuery.isPending ? (
            <div className="flex items-center gap-2 px-4 py-3 text-sm text-brand-text-muted">
              <Loader2 aria-hidden="true" className="size-4 animate-spin" />
              Searching
            </div>
          ) : results.length === 0 ? (
            <p className="px-4 py-3 text-sm text-brand-text-muted">No users match that name.</p>
          ) : (
            <ul className={usersQuery.isFetching ? "opacity-70 transition-opacity" : undefined}>
              {results.map((user) => {
                const isAssigned = assignedIDs.has(user.id);
                const isAssigning =
                  assignRole.isPending && assignRole.variables?.userId === user.id;

                return (
                  <li
                    key={user.id}
                    className="flex items-center gap-3 border-b border-brand-border/70 px-4 py-2.5 last:border-b-0"
                  >
                    <Avatar src={user.avatar_url} name={user.full_name} email={user.email} />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-brand-text">
                        {user.full_name}
                      </p>
                      <p className="truncate text-xs text-brand-text-subtle">{user.email}</p>
                    </div>
                    {isAssigned ? (
                      <span className="text-xs text-brand-text-subtle">Has role</span>
                    ) : (
                      <ActionButton
                        onClick={() => handleAssign(user.id, user.full_name)}
                        loading={isAssigning}
                        disabled={assignRole.isPending}
                        icon={<UserPlus aria-hidden="true" className="size-4" />}
                        loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
                        className="h-9 px-3"
                      >
                        Add
                      </ActionButton>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      ) : null}
    </div>
  );
}

type RoleUsersPanelProps = {
  role: RoleDetails;
};

export function RoleUsersPanel({ role }: RoleUsersPanelProps) {
  const [confirmingID, setConfirmingID] = useState<string | null>(null);
  const unassignRole = useUnassignRole();
  const users = role.users_with_role;

  const handleUnassign = async (userId: string, name: string) => {
    try {
      await unassignRole.mutateAsync({ roleId: role.id, userId });
      setConfirmingID(null);
      toast.success(`Removed ${role.name} from ${name}`);
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <SurfacePanel className="bg-brand-surface/80">
      <div className="border-b border-brand-border px-5 py-5 sm:px-6">
        <h2 className="text-base font-semibold text-brand-text">Users with this role</h2>
        <p className="mt-1 text-sm text-brand-text-subtle">
          <span className="font-mono text-brand-text">{users.length}</span>{" "}
          {users.length === 1 ? "user has" : "users have"} this role.
        </p>
      </div>

      <RoleUserSearch role={role} />

      {users.length === 0 ? (
        <p className="px-5 py-8 text-center text-sm text-brand-text-muted sm:px-6">
          No one has this role yet. Search above to add someone.
        </p>
      ) : (
        <ul>
          {users.map((user) => {
            const isConfirming = confirmingID === user.user_id;
            const isRemoving =
              unassignRole.isPending && unassignRole.variables?.userId === user.user_id;

            return (
              <li
                key={user.user_id}
                className="flex flex-wrap items-center gap-3 border-b border-brand-border/70 px-5 py-3 last:border-b-0 sm:px-6"
              >
                <Avatar src={user.profile_picture} name={user.name} />
                <Link
                  href={`/admin/users/${user.user_id}`}
                  className="min-w-0 flex-1 truncate text-sm font-medium text-brand-text hover:underline"
                >
                  {user.name}
                </Link>
                {isConfirming ? (
                  <div className="flex items-center gap-2">
                    <ActionButton
                      aria-label={`Keep ${role.name} for ${user.name}`}
                      onClick={() => setConfirmingID(null)}
                      disabled={isRemoving}
                      icon={<X aria-hidden="true" className="size-4" />}
                      className="h-9 px-3"
                    >
                      Cancel
                    </ActionButton>
                    <ActionButton
                      onClick={() => handleUnassign(user.user_id, user.name)}
                      loading={isRemoving}
                      icon={<UserMinus aria-hidden="true" className="size-4" />}
                      loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
                      className="h-9 border-red-400/50 bg-red-400/10 px-3 text-red-100 hover:border-red-300 hover:bg-red-400/20"
                    >
                      Confirm remove
                    </ActionButton>
                  </div>
                ) : (
                  <ActionButton
                    aria-label={`Remove ${role.name} from ${user.name}`}
                    onClick={() => setConfirmingID(user.user_id)}
                    disabled={unassignRole.isPending}
                    icon={<UserMinus aria-hidden="true" className="size-4" />}
                    className="h-9 border-red-400/30 px-3 text-red-200 hover:border-red-300 hover:bg-red-400/10"
                  >
                    Remove
                  </ActionButton>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </SurfacePanel>
  );
}
