import { cn } from "@/lib/utils/cn";

type CheckboxOption<T extends string> = {
  value: T;
  label: string;
  description?: string;
};

type CheckboxFilterProps<T extends string> = {
  label: string;
  options: CheckboxOption<T>[];
  selectedValues: T[];
  onChange: (values: T[]) => void;
  allLabel?: string;
  helpText?: string;
};

/**
 * Renders a controlled checkbox filter with optional help text.
 * An empty selection represents all options.
 */
export function CheckboxFilter<T extends string>({
  label,
  options,
  selectedValues,
  onChange,
  allLabel = "All",
  helpText = "Selected values must match the user's complete set.",
}: CheckboxFilterProps<T>) {
  const isAllSelected = selectedValues.length === 0;

  const toggleValue = (value: T, checked: boolean) => {
    if (checked) {
      onChange([...selectedValues, value]);
      return;
    }

    onChange(selectedValues.filter((selectedValue) => selectedValue !== value));
  };

  return (
    <fieldset className="min-w-0">
      <legend className="mb-1.5 text-sm text-brand-text-subtle">{label}</legend>
      <div className="flex max-h-32 flex-wrap gap-2 overflow-y-auto border border-brand-border bg-brand-surface p-2">
        <label
          className={cn(
            "flex min-h-9 cursor-pointer items-center gap-2 border px-3 text-sm transition-colors",
            isAllSelected
              ? "border-brand-primary bg-brand-primary/20 text-brand-text"
              : "border-brand-border bg-white/2 text-brand-text-muted hover:bg-white/5",
          )}
        >
          <input
            type="checkbox"
            checked={isAllSelected}
            onChange={() => onChange([])}
            className="size-4 accent-brand-primary"
          />
          <span>{allLabel}</span>
        </label>

        {options.map((option) => {
          const checked = selectedValues.includes(option.value);

          return (
            <label
              key={option.value}
              className={cn(
                "flex min-h-9 cursor-pointer items-center gap-2 border px-3 text-sm transition-colors",
                checked
                  ? "border-brand-primary bg-brand-primary/20 text-brand-text"
                  : "border-brand-border bg-white/2 text-brand-text-muted hover:bg-white/5",
              )}
            >
              <input
                type="checkbox"
                checked={checked}
                onChange={(event) => toggleValue(option.value, event.target.checked)}
                className="size-4 accent-brand-primary"
              />
              <span>{option.label}</span>
              {option.description && (
                <span className="text-xs text-brand-text-subtle">{option.description}</span>
              )}
            </label>
          );
        })}
      </div>
      <p className="mt-1.5 text-xs text-brand-text-subtle">{helpText}</p>
    </fieldset>
  );
}
