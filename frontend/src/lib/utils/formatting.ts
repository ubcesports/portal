export const TIME_FORMAT = new Intl.DateTimeFormat("en", {
  month: "short",
  day: "numeric",
  year: "numeric",
  hour: "numeric",
  minute: "2-digit",
});

export const DATE_FORMAT = new Intl.DateTimeFormat("en", {
  month: "short",
  day: "numeric",
  year: "numeric",
});

export const formatTime = (date: number | string | Date): string =>
  TIME_FORMAT.format(new Date(date));

export const formatDate = (date: number | string | Date): string =>
  DATE_FORMAT.format(new Date(date));

const CURRENCY_FORMAT = new Intl.NumberFormat("en-CA", {
  style: "currency",
  currency: "CAD",
  maximumFractionDigits: 0,
});

const COMPACT_NUMBER_FORMAT = new Intl.NumberFormat("en", {
  notation: "compact",
  maximumFractionDigits: 1,
});

const NUMBER_FORMAT = new Intl.NumberFormat("en");

/** Formats cents as Canadian dollars rounded to whole dollars using the en-CA locale. */
export const formatCentsAsCurrency = (cents: number): string => CURRENCY_FORMAT.format(cents / 100);

/** Formats a number in English compact notation with at most one fractional digit. */
export const formatCompactNumber = (value: number): string => COMPACT_NUMBER_FORMAT.format(value);

/** Formats a number with English grouping and default Intl fractional precision. */
export const formatNumber = (value: number): string => NUMBER_FORMAT.format(value);

export function getInitials(name: string, email: string): string {
  const source = name && name !== "Profile" ? name : email;
  const parts = source
    .split(/[\s@.]+/)
    .map((part) => part.trim())
    .filter(Boolean);

  if (parts.length === 0) {
    return "UB";
  }

  return parts
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}
