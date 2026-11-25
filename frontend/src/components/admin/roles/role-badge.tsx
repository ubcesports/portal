type RoleBadgeProps = {
  name: string;
  color: string | null;
};

/* Shows a role's name with its color swatch. The color is user-defined, so it is set inline. */
export function RoleBadge({ name, color }: RoleBadgeProps) {
  return (
    <span className="inline-flex min-h-6 max-w-full items-center gap-2 border border-brand-border bg-white/5 px-2 text-xs font-semibold text-brand-text">
      <span
        aria-hidden="true"
        className="size-2.5 shrink-0 rounded-full border border-white/20"
        style={{ backgroundColor: color ?? "transparent" }}
      />
      <span className="truncate">{name}</span>
    </span>
  );
}
