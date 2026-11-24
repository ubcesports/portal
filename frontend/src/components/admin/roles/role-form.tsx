"use client";

import { Loader2, Save } from "lucide-react";
import { useId, useState, type FormEvent, type ReactNode } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { RolePermissionsPicker } from "@/components/admin/roles/role-permissions-picker";
import { SurfacePanel } from "@/components/surface-panel";
import { cn } from "@/lib/utils/cn";
import type { PermissionKey } from "@/lib/types/permission.types";
import {
  ROLE_NAME_MAX_LENGTH,
  type CreateRoleRequest,
  type RoleDetails,
  type UpdateRoleRequest,
} from "@/lib/types/role.types";

const FIELD_CLASS_NAME =
  "h-11 w-full border border-brand-border bg-brand-surface px-3 text-sm text-brand-text outline-none transition placeholder:text-brand-text-subtle focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:cursor-not-allowed disabled:opacity-60";

const DEFAULT_COLOR = "#5865F2";
const COLOR_PATTERN = /^#[0-9A-Fa-f]{6}$/;

type Draft = {
  name: string;
  description: string;
  hasColor: boolean;
  color: string;
  permissionKeys: Set<PermissionKey>;
};

function toDraft(role?: RoleDetails): Draft {
  return {
    name: role?.name ?? "",
    description: role?.description ?? "",
    hasColor: !!role?.color,
    color: role?.color ?? DEFAULT_COLOR,
    permissionKeys: new Set(role?.permission_keys ?? []),
  };
}

function sameKeys(a: ReadonlySet<PermissionKey>, b: readonly PermissionKey[]) {
  return a.size === new Set(b).size && b.every((key) => a.has(key));
}

/* Returns an error message, or null when the draft is valid. */
function validate(draft: Draft): string | null {
  const name = draft.name.trim();
  if (!name) {
    return "Enter a role name";
  }
  // Count characters, not UTF-16 units, to match the API.
  if ([...name].length > ROLE_NAME_MAX_LENGTH) {
    return `Role names can be at most ${ROLE_NAME_MAX_LENGTH} characters`;
  }
  if (draft.hasColor && !COLOR_PATTERN.test(draft.color)) {
    return "Enter a color like #5865F2";
  }
  return null;
}

function buildCreateBody(draft: Draft): CreateRoleRequest {
  const body: CreateRoleRequest = {
    name: draft.name.trim(),
    permission_keys: [...draft.permissionKeys],
  };
  const description = draft.description.trim();
  if (description) {
    body.description = description;
  }
  if (draft.hasColor) {
    body.color = draft.color.toUpperCase();
  }
  return body;
}

/* Builds the PATCH body from the fields that changed. Absent fields are left untouched. */
function buildUpdateBody(role: RoleDetails, draft: Draft): UpdateRoleRequest {
  const body: UpdateRoleRequest = {};
  const name = draft.name.trim();
  if (name !== role.name) {
    body.name = name;
  }
  const description = draft.description.trim();
  if (description !== (role.description ?? "")) {
    body.description = description;
  }
  const color = draft.color.toUpperCase();
  if (draft.hasColor && color !== role.color?.toUpperCase()) {
    body.color = color;
  }
  if (!sameKeys(draft.permissionKeys, role.permission_keys)) {
    body.permission_keys = [...draft.permissionKeys];
  }
  return body;
}

type RoleFormProps = {
  title: string;
  description: string;
  isSubmitting: boolean;
  submitLabel: string;
  footer?: ReactNode;
} & (
  | { role?: undefined; onSubmit: (body: CreateRoleRequest) => Promise<unknown> }
  | { role: RoleDetails; onSubmit: (body: UpdateRoleRequest) => Promise<unknown> }
);

/*
  Creates a role, or edits one when `role` is given. The parent should key this
  component by the role's updated_at so the draft resets after a save.
*/
export function RoleForm(props: RoleFormProps) {
  const { title, description, isSubmitting, submitLabel, footer } = props;
  const formID = useId();
  const [draft, setDraft] = useState(() => toDraft(props.role));

  const update = <K extends keyof Draft>(key: K, value: Draft[K]) =>
    setDraft((current) => ({ ...current, [key]: value }));

  // The API can set a color but not clear one, so an existing color stays on.
  const canRemoveColor = !props.role?.color;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const error = validate(draft);
    if (error) {
      toast.error(error);
      return;
    }

    try {
      if (props.role) {
        const body = buildUpdateBody(props.role, draft);
        if (Object.keys(body).length === 0) {
          toast.info("No changes to save");
          return;
        }
        await props.onSubmit(body);
      } else {
        await props.onSubmit(buildCreateBody(draft));
      }
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <SurfacePanel className="bg-brand-surface/80">
      <form onSubmit={handleSubmit}>
        <div className="border-b border-brand-border px-5 py-5 sm:px-6">
          <h1 className="text-lg font-semibold text-brand-text">{title}</h1>
          <p className="mt-1 text-sm text-brand-text-subtle">{description}</p>
        </div>

        <fieldset disabled={isSubmitting} className="flex flex-col gap-6 px-5 py-5 sm:px-6">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
              <span>Name</span>
              <input
                id={`${formID}-name`}
                required
                value={draft.name}
                onChange={(event) => update("name", event.target.value)}
                placeholder="Lounge staff"
                className={FIELD_CLASS_NAME}
              />
            </label>

            <div className="flex flex-col gap-1.5 text-sm text-brand-text-subtle">
              <label htmlFor={`${formID}-color`}>Color</label>
              <div className="flex h-11 items-center gap-2">
                <input
                  id={`${formID}-color`}
                  type="color"
                  value={draft.color}
                  // Picking the color already shown fires no change event.
                  onClick={() => update("hasColor", true)}
                  onChange={(event) => {
                    update("color", event.target.value);
                    update("hasColor", true);
                  }}
                  className={`h-11 w-14 shrink-0 cursor-pointer border border-brand-border bg-brand-surface p-1 ${
                    draft.hasColor ? "" : "opacity-40"
                  }`}
                />
                {draft.hasColor ? (
                  <span className="font-mono text-xs text-brand-text">
                    {draft.color.toUpperCase()}
                  </span>
                ) : (
                  <span className="text-xs text-brand-text-subtle">No color</span>
                )}
                {canRemoveColor && draft.hasColor ? (
                  <button
                    type="button"
                    onClick={() => update("hasColor", false)}
                    className="ml-auto cursor-pointer text-xs text-brand-text-subtle underline-offset-2 hover:text-brand-text hover:underline"
                  >
                    Remove
                  </button>
                ) : null}
              </div>
            </div>

            <label className="flex flex-col gap-1.5 text-sm text-brand-text-subtle md:col-span-2">
              <span>Description</span>
              <textarea
                id={`${formID}-description`}
                rows={2}
                value={draft.description}
                onChange={(event) => update("description", event.target.value)}
                placeholder="What this role is for"
                className={cn(FIELD_CLASS_NAME, "h-auto py-2.5")}
              />
            </label>
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex items-baseline justify-between gap-3">
              <h2 className="text-sm font-semibold text-brand-text">Permissions</h2>
              <span className="text-xs text-brand-text-subtle">
                <span className="font-mono text-brand-text">{draft.permissionKeys.size}</span>{" "}
                selected
              </span>
            </div>
            <RolePermissionsPicker
              selected={draft.permissionKeys}
              onChange={(keys) => update("permissionKeys", keys)}
            />
          </div>
        </fieldset>

        <div className="flex flex-wrap items-center justify-between gap-3 border-t border-brand-border px-5 py-4 sm:px-6">
          <div>{footer}</div>
          <ActionButton
            type="submit"
            loading={isSubmitting}
            icon={<Save aria-hidden="true" className="size-4" />}
            loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            className="h-11 border-brand-primary bg-brand-primary px-5 text-white hover:border-brand-primary-hover hover:bg-brand-primary-hover"
          >
            {submitLabel}
          </ActionButton>
        </div>
      </form>
    </SurfacePanel>
  );
}
