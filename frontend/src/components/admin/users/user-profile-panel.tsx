"use client";

import { Loader2, Pencil, Save, X } from "lucide-react";
import Image from "next/image";
import { useState } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { DetailRow } from "@/components/detail-row";
import { StatusBadge } from "@/components/status-badge";
import { SurfacePanel } from "@/components/surface-panel";
import type { GroupType, RoleType, User } from "@/lib/types/user.types";
import type { AdminUserProfile, UpdateUserRequest } from "@/lib/types/admin.types";
import { GROUP_OPTIONS, ROLE_OPTIONS } from "@/lib/types/admin.types";
import { formatTime, getInitials } from "@/lib/utils/formatting";
import { getGroupBadgeClass, titleCase } from "@/lib/utils/groups";

const FIELD_CLASS_NAME =
  "h-10 border border-brand-border bg-brand-surface px-3 text-sm text-brand-text";

const STUDENT_ID_PATTERN = /^\d{8}$/;

type UserProfilePanelProps = {
  user: AdminUserProfile;
  onSave: (body: UpdateUserRequest) => Promise<unknown>;
  isSaving: boolean;
};

export type Draft = {
  fullName: string;
  isStudent: boolean;
  studentId: string;
  role: RoleType;
  groups: GroupType[];
};

function toDraft(user: User): Draft {
  return {
    fullName: user.full_name,
    isStudent: user.is_student,
    studentId: user.student_id ?? "",
    role: user.role,
    groups: [...user.groups],
  };
}

function EmptyValue() {
  return <span className="text-brand-text-muted">—</span>;
}

function OptionalTime({ value }: { value: string | null }) {
  if (!value) {
    return <EmptyValue />;
  }

  return <>{formatTime(value)}</>;
}

/*
  Builds the PATCH body from the fields that actually changed. Absent fields are
  left untouched by the API, so an unchanged field must not be sent — notably
  `student_id`, which the API rejects for non-students.
*/
export function buildUpdateBody(user: User, draft: Draft): UpdateUserRequest {
  const body: UpdateUserRequest = {};
  const trimmedFullName = draft.fullName.trim();
  if (trimmedFullName !== user.full_name) {
    body.full_name = trimmedFullName;
  }
  const trimmedStudentId = draft.studentId.trim();

  if (draft.isStudent !== user.is_student) {
    body.is_student = draft.isStudent;

    // Becoming a student needs an ID; dropping student status must not send one,
    // because the API generates the replacement itself.
    if (draft.isStudent) {
      body.student_id = trimmedStudentId;
    }
  } else if (draft.isStudent && trimmedStudentId !== (user.student_id ?? "")) {
    body.student_id = trimmedStudentId;
  }

  if (draft.role !== user.role) {
    body.role = draft.role;
  }

  const groupsAdd = draft.groups.filter((group) => !user.groups.includes(group));
  // The member group is permanent, so it is never offered for removal.
  const groupsRemove = user.groups.filter(
    (group) => group !== "member" && !draft.groups.includes(group),
  );

  if (groupsAdd.length > 0) {
    body.groups_add = groupsAdd;
  }
  if (groupsRemove.length > 0) {
    body.groups_remove = groupsRemove;
  }

  return body;
}

export function validateDraft(draft: Draft): string | null {
  if (!draft.fullName.trim()) {
    return "Full name is required.";
  }
  if (draft.isStudent && !STUDENT_ID_PATTERN.test(draft.studentId.trim())) {
    return "Student ID must be an 8 digit number.";
  }

  return null;
}

export function UserProfilePanel({ user, onSave, isSaving }: UserProfilePanelProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState<Draft>(() => toDraft(user));

  const startEditing = () => {
    setDraft(toDraft(user));
    setIsEditing(true);
  };

  const cancelEditing = () => {
    setDraft(toDraft(user));
    setIsEditing(false);
  };

  const handleSave = async () => {
    const validationError = validateDraft(draft);
    if (validationError) {
      toast.error(validationError);
      return;
    }

    const body = buildUpdateBody(user, draft);
    if (Object.keys(body).length === 0) {
      setIsEditing(false);
      return;
    }

    try {
      await onSave(body);
      toast.success("User updated");
      setIsEditing(false);
    } catch {
      // The API client already surfaces the error message as a toast.
    }
  };

  const toggleGroup = (group: GroupType) => {
    setDraft((current) => ({
      ...current,
      groups: current.groups.includes(group)
        ? current.groups.filter((value) => value !== group)
        : [...current.groups, group],
    }));
  };

  return (
    <SurfacePanel className="bg-transparent">
      <div className="flex flex-col gap-4 border-b border-brand-border px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          {user.avatar_url ? (
            <Image
              src={user.avatar_url}
              alt=""
              width={44}
              height={44}
              className="size-11 border border-brand-border object-cover"
              unoptimized
            />
          ) : (
            <span className="flex size-11 items-center justify-center border border-brand-border bg-brand-primary/20 text-sm font-bold text-brand-text">
              {getInitials(user.full_name, user.email)}
            </span>
          )}
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold text-brand-text">{user.full_name}</h2>
            <p className="truncate text-sm text-brand-text-subtle">{user.email}</p>
          </div>
        </div>

        {isEditing ? (
          <div className="flex shrink-0 gap-2">
            <ActionButton
              onClick={handleSave}
              loading={isSaving}
              icon={<Save aria-hidden="true" className="size-4" />}
              loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            >
              {isSaving ? "Saving" : "Save"}
            </ActionButton>
            <ActionButton
              onClick={cancelEditing}
              disabled={isSaving}
              icon={<X aria-hidden="true" className="size-4" />}
            >
              Cancel
            </ActionButton>
          </div>
        ) : (
          <ActionButton
            onClick={startEditing}
            className="shrink-0"
            icon={<Pencil aria-hidden="true" className="size-4" />}
          >
            Edit
          </ActionButton>
        )}
      </div>

      <dl>
        <DetailRow label="Full name">
          {isEditing ? (
            <input
              type="text"
              aria-label="Full name"
              value={draft.fullName}
              disabled={isSaving}
              onChange={(event) =>
                setDraft((current) => ({ ...current, fullName: event.target.value }))
              }
              className={`${FIELD_CLASS_NAME} w-full max-w-sm`}
            />
          ) : (
            user.full_name
          )}
        </DetailRow>

        <DetailRow label="Admin note">
          {user.admin_note ? (
            <span className="whitespace-pre-wrap break-words">{user.admin_note}</span>
          ) : (
            <EmptyValue />
          )}
        </DetailRow>

        <DetailRow label="Student">
          {isEditing ? (
            <label className="flex items-center gap-2 text-sm text-brand-text">
              <input
                type="checkbox"
                checked={draft.isStudent}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    isStudent: event.target.checked,
                    // A generated ID replaces the old one, so clear the field to
                    // make that visible before saving.
                    studentId: event.target.checked ? current.studentId : "",
                  }))
                }
                className="size-4 accent-brand-primary"
              />
              <span>Is a UBC student</span>
            </label>
          ) : (
            <StatusBadge tone={user.is_student ? "success" : "muted"}>
              {user.is_student ? "Student" : "Non-student"}
            </StatusBadge>
          )}
        </DetailRow>

        <DetailRow label="Student ID">
          {isEditing ? (
            <div>
              <input
                type="text"
                inputMode="numeric"
                value={draft.studentId}
                disabled={!draft.isStudent}
                placeholder={draft.isStudent ? "12345678" : "Generated automatically"}
                onChange={(event) =>
                  setDraft((current) => ({ ...current, studentId: event.target.value }))
                }
                className={`${FIELD_CLASS_NAME} w-48 disabled:opacity-60`}
              />
              {!draft.isStudent ? (
                <p className="mt-1 text-xs text-brand-text-muted">
                  A non-student ID is generated on save.
                </p>
              ) : null}
            </div>
          ) : user.student_id ? (
            user.student_id
          ) : (
            <EmptyValue />
          )}
        </DetailRow>

        <DetailRow label="Role">
          {isEditing ? (
            <select
              value={draft.role}
              onChange={(event) =>
                setDraft((current) => ({ ...current, role: event.target.value as RoleType }))
              }
              className={`${FIELD_CLASS_NAME} w-48`}
            >
              {ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          ) : (
            <StatusBadge tone={user.role === "admin" ? "warning" : "muted"}>
              {titleCase(user.role)}
            </StatusBadge>
          )}
        </DetailRow>

        <DetailRow label="Groups">
          {isEditing ? (
            <div className="flex flex-wrap gap-2">
              {GROUP_OPTIONS.map((option) => {
                const isSelected = draft.groups.includes(option.value);
                // Membership in the member group is permanent.
                const isLocked = option.value === "member";

                return (
                  <button
                    key={option.value}
                    type="button"
                    disabled={isLocked}
                    onClick={() => toggleGroup(option.value)}
                    title={isLocked ? "Every user is a member" : undefined}
                    className={`inline-flex min-h-7 items-center border px-2 text-xs font-semibold transition ${
                      isSelected
                        ? getGroupBadgeClass(option.value)
                        : "border-brand-border bg-transparent text-brand-text-muted hover:border-brand-text-muted"
                    } ${isLocked ? "cursor-not-allowed opacity-70" : "cursor-pointer"}`}
                  >
                    {option.label}
                  </button>
                );
              })}
            </div>
          ) : user.groups.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {user.groups.map((group) => (
                <StatusBadge key={group} className={getGroupBadgeClass(group)}>
                  {titleCase(group)}
                </StatusBadge>
              ))}
            </div>
          ) : (
            <EmptyValue />
          )}
        </DetailRow>

        <DetailRow label="Email verified">
          <OptionalTime value={user.email_verified_at} />
        </DetailRow>
        <DetailRow label="Onboarded">
          <OptionalTime value={user.onboarding_completed_at} />
        </DetailRow>
        <DetailRow label="Created">{formatTime(user.created_at)}</DetailRow>
        <DetailRow label="Updated">{formatTime(user.updated_at)}</DetailRow>
        <DetailRow label="User ID">
          <span className="break-all font-mono text-xs text-brand-text-muted">{user.id}</span>
        </DetailRow>
      </dl>
    </SurfacePanel>
  );
}
