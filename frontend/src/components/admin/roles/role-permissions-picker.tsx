"use client";

import { PERMISSION_GROUPS, type PermissionKey } from "@/lib/types/permission.types";

type RolePermissionsPickerProps = {
  selected: ReadonlySet<PermissionKey>;
  onChange: (selected: Set<PermissionKey>) => void;
  disabled?: boolean;
};

export function RolePermissionsPicker({
  selected,
  onChange,
  disabled,
}: RolePermissionsPickerProps) {
  const toggle = (keys: PermissionKey[], checked: boolean) => {
    const next = new Set(selected);
    for (const key of keys) {
      if (checked) {
        next.add(key);
      } else {
        next.delete(key);
      }
    }
    onChange(next);
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {PERMISSION_GROUPS.map((group) => {
        const keys = group.permissions.map((permission) => permission.key);
        const selectedCount = keys.filter((key) => selected.has(key)).length;
        const allSelected = selectedCount === keys.length;

        return (
          <fieldset
            key={group.label}
            disabled={disabled}
            className="border border-brand-border bg-white/2 disabled:opacity-60"
          >
            <legend className="sr-only">{group.label}</legend>
            <div className="flex items-center justify-between gap-3 border-b border-brand-border px-4 py-2.5">
              <span className="text-xs font-semibold uppercase tracking-wide text-brand-text-subtle">
                {group.label}
              </span>
              <label className="flex cursor-pointer items-center gap-2 text-xs text-brand-text-subtle">
                <input
                  type="checkbox"
                  checked={allSelected}
                  ref={(input) => {
                    if (input) {
                      input.indeterminate = selectedCount > 0 && !allSelected;
                    }
                  }}
                  onChange={(event) => toggle(keys, event.target.checked)}
                  className="size-4 accent-brand-primary"
                />
                All
              </label>
            </div>
            <div className="flex flex-col py-1">
              {group.permissions.map((permission) => (
                <label
                  key={permission.key}
                  className="flex cursor-pointer items-start gap-3 px-4 py-2 transition hover:bg-white/3"
                >
                  <input
                    type="checkbox"
                    checked={selected.has(permission.key)}
                    onChange={(event) => toggle([permission.key], event.target.checked)}
                    className="mt-0.5 size-4 shrink-0 accent-brand-primary"
                  />
                  <span>
                    <span className="block text-sm text-brand-text">{permission.label}</span>
                    <span className="block text-xs text-brand-text-subtle">
                      {permission.description}
                    </span>
                  </span>
                </label>
              ))}
            </div>
          </fieldset>
        );
      })}
    </div>
  );
}
