"use client";

import { Loader2, Search, X } from "lucide-react";
import Link from "next/link";
import { useId, useRef, useState, type KeyboardEvent } from "react";
import { toast } from "sonner";
import { SurfacePanel } from "@/components/surface-panel";
import {
  useAssignRole,
  useRoles,
  useUnassignRole,
  useUserRoles,
} from "@/lib/admin/admin.roles.hook";
import type { Role, UserRole } from "@/lib/types/role.types";
import type { User } from "@/lib/types/user.types";

function RoleDot({ color }: { color: string | null }) {
  return (
    <span
      aria-hidden="true"
      className="size-2.5 shrink-0 rounded-full border border-white/20"
      style={{ backgroundColor: color ?? "transparent" }}
    />
  );
}

type UserCustomRolesPanelProps = {
  user: User;
};

/*
  Lets an admin give custom roles to one user by searching for them, and take
  them away from the chips above the search box.
*/
export function UserCustomRolesPanel({ user }: UserCustomRolesPanelProps) {
  const listboxID = useId();
  const inputRef = useRef<HTMLInputElement>(null);
  const [search, setSearch] = useState("");
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

  const rolesQuery = useRoles();
  const userRolesQuery = useUserRoles(user.id);
  const roles = rolesQuery.data ?? [];
  const heldRoles = userRolesQuery.data ?? [];
  const assignRole = useAssignRole();
  const unassignRole = useUnassignRole();

  const isLoading = rolesQuery.isPending || userRolesQuery.isPending;
  const hasError = rolesQuery.isError || userRolesQuery.isError;
  const isMutating = assignRole.isPending || unassignRole.isPending;

  const heldRoleIDs = new Set(heldRoles.map((role) => role.id));

  const query = search.trim().toLowerCase();
  const matches = roles.filter(
    (role) => !heldRoleIDs.has(role.id) && role.name.toLowerCase().includes(query),
  );
  const highlighted = matches[Math.min(activeIndex, matches.length - 1)];

  const giveRole = async (role: Role) => {
    try {
      await assignRole.mutateAsync({ roleId: role.id, userId: user.id });
      toast.success(`Gave ${role.name} to ${user.full_name}`);
      setSearch("");
      setActiveIndex(0);
      inputRef.current?.focus();
    } catch {
      // The shared API client displays the server error.
    }
  };

  const removeRole = async (role: UserRole) => {
    try {
      await unassignRole.mutateAsync({ roleId: role.id, userId: user.id });
      toast.success(`Removed ${role.name} from ${user.full_name}`);
    } catch {
      // The shared API client displays the server error.
    }
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setIsOpen(true);
      setActiveIndex((index) => Math.min(index + 1, matches.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((index) => Math.max(index - 1, 0));
    } else if (event.key === "Enter") {
      event.preventDefault();
      if (isOpen && highlighted && !isMutating) {
        void giveRole(highlighted);
      }
    } else if (event.key === "Escape") {
      setIsOpen(false);
    }
  };

  return (
    <SurfacePanel className="bg-transparent">
      <div className="border-b border-brand-border px-5 py-4">
        <h2 className="text-base font-semibold text-brand-text">Custom roles</h2>
        <p className="mt-1 text-sm text-brand-text-subtle">
          {user.role === "admin"
            ? "Admins already have every permission, so custom roles only take effect if this user becomes a member."
            : "Search to give this user a role. Changes save immediately."}
        </p>
      </div>

      <div className="flex flex-col gap-3 px-5 py-4">
        {isLoading ? (
          <div className="flex items-center gap-2 text-sm text-brand-text-muted">
            <Loader2 aria-hidden="true" className="size-4 animate-spin" />
            Loading roles
          </div>
        ) : hasError ? (
          <p className="text-sm text-brand-text-muted">Unable to load roles.</p>
        ) : roles.length === 0 ? (
          <p className="text-sm text-brand-text-muted">
            No custom roles yet.{" "}
            <Link
              href="/admin/roles/new"
              className="text-brand-text underline-offset-2 hover:underline"
            >
              Create one
            </Link>
            .
          </p>
        ) : (
          <>
            {heldRoles.length > 0 ? (
              <ul aria-label="Roles this user has" className="flex flex-wrap gap-2">
                {heldRoles.map((role) => {
                  const isRemoving =
                    unassignRole.isPending && unassignRole.variables?.roleId === role.id;

                  return (
                    <li
                      key={role.id}
                      className="inline-flex min-h-8 items-center gap-2 border border-brand-primary/60 bg-brand-primary/15 pl-2.5 text-xs font-semibold text-brand-text"
                    >
                      <RoleDot color={role.color} />
                      {role.name}
                      <button
                        type="button"
                        aria-label={`Remove ${role.name}`}
                        title={`Remove ${role.name}`}
                        disabled={isMutating}
                        onClick={() => removeRole(role)}
                        className="flex h-8 w-7 cursor-pointer items-center justify-center text-brand-text-subtle transition hover:bg-white/5 hover:text-brand-text disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {isRemoving ? (
                          <Loader2 aria-hidden="true" className="size-3.5 animate-spin" />
                        ) : (
                          <X aria-hidden="true" className="size-3.5" />
                        )}
                      </button>
                    </li>
                  );
                })}
              </ul>
            ) : (
              <p className="text-sm text-brand-text-muted">This user has no custom roles.</p>
            )}

            <div className="relative max-w-md">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-brand-text-subtle"
              />
              <input
                ref={inputRef}
                type="text"
                role="combobox"
                aria-label="Add a role"
                aria-expanded={isOpen}
                aria-controls={listboxID}
                aria-autocomplete="list"
                aria-activedescendant={
                  isOpen && highlighted ? `${listboxID}-${highlighted.id}` : undefined
                }
                value={search}
                placeholder="Add a role…"
                onChange={(event) => {
                  setSearch(event.target.value);
                  setActiveIndex(0);
                  setIsOpen(true);
                }}
                onFocus={() => setIsOpen(true)}
                // Delay so a click on an option lands before the list closes.
                onBlur={() => setTimeout(() => setIsOpen(false), 150)}
                onKeyDown={handleKeyDown}
                className="h-10 w-full border border-brand-border bg-brand-surface pl-9 pr-9 text-sm text-brand-text outline-none transition placeholder:text-brand-text-subtle focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:opacity-60"
              />
              {assignRole.isPending ? (
                <Loader2
                  aria-hidden="true"
                  className="absolute right-3 top-1/2 size-4 -translate-y-1/2 animate-spin text-brand-text-subtle"
                />
              ) : null}

              {isOpen ? (
                <ul
                  id={listboxID}
                  role="listbox"
                  className="absolute inset-x-0 top-[calc(100%+4px)] z-20 max-h-64 overflow-auto border border-brand-border bg-brand-surface shadow-2xl shadow-black/40"
                >
                  {matches.length === 0 ? (
                    <li className="px-3 py-2.5 text-sm text-brand-text-muted">
                      {roles.every((role) => heldRoleIDs.has(role.id))
                        ? "This user already has every role."
                        : "No roles match."}
                    </li>
                  ) : (
                    matches.map((role) => (
                      <li
                        key={role.id}
                        id={`${listboxID}-${role.id}`}
                        role="option"
                        aria-selected={role.id === highlighted?.id}
                        // Keep focus in the input so the list stays open.
                        onMouseDown={(event) => event.preventDefault()}
                        onMouseEnter={() => setActiveIndex(matches.indexOf(role))}
                        onClick={() => !isMutating && giveRole(role)}
                        className={`flex cursor-pointer items-center gap-2 px-3 py-2.5 text-sm text-brand-text ${
                          role.id === highlighted?.id ? "bg-white/5" : ""
                        }`}
                      >
                        <RoleDot color={role.color} />
                        <span className="truncate">{role.name}</span>
                      </li>
                    ))
                  )}
                </ul>
              ) : null}
            </div>
          </>
        )}
      </div>
    </SurfacePanel>
  );
}
