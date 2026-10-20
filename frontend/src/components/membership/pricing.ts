import type {
  EligibleMembershipTier,
  MembershipExpirationType,
  MembershipTier,
  MembershipTierPrice,
  TierUnavailableReason,
} from "@/lib/types/membership.types";

const STUDENT_LABELS: Record<string, string> = {
  true: "Student",
  false: "Community",
  null: "Standard",
};

const VANCOUVER_TIME_ZONE = "America/Vancouver";

const VANCOUVER_DATE_FORMATTER = new Intl.DateTimeFormat("en-CA", {
  timeZone: VANCOUVER_TIME_ZONE,
  month: "long",
  day: "numeric",
  year: "numeric",
});

const VANCOUVER_DATE_PARTS_FORMATTER = new Intl.DateTimeFormat("en-CA", {
  timeZone: VANCOUVER_TIME_ZONE,
  month: "numeric",
  day: "numeric",
  year: "numeric",
});

export function formatMembershipPrice(price: number) {
  return new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    minimumFractionDigits: Number.isInteger(price) ? 0 : 2,
    maximumFractionDigits: 2,
  }).format(price);
}

export function formatMembershipExpiration(
  expirationType: MembershipExpirationType,
  now = new Date(),
) {
  const dateParts = Object.fromEntries(
    VANCOUVER_DATE_PARTS_FORMATTER.formatToParts(now).map(({ type, value }) => [type, value]),
  );
  const currentMonth = Number(dateParts.month);
  const currentYear = Number(dateParts.year);

  switch (expirationType) {
    case "day":
      return `Valid until the end of the day (${VANCOUVER_DATE_FORMATTER.format(now)})`;

    case "semester": {
      const isWinterSemester = currentMonth <= 4;
      const expirationMonth = isWinterSemester ? "April" : "December";
      const expirationDay = isWinterSemester ? 30 : 31;

      return `Valid until the end of the semester (${expirationMonth} ${expirationDay}, ${currentYear})`;
    }

    case "year": {
      const expirationYear = currentMonth >= 5 ? currentYear + 1 : currentYear;

      return `Valid until the end of the school year (April 30, ${expirationYear})`;
    }

    default:
      return "Expiration date unavailable";
  }
}

export function membershipPriceLabel(price: MembershipTierPrice) {
  return STUDENT_LABELS[String(price.is_student_required)] ?? "Member";
}

export function getPriceByStudentStatus(tier: MembershipTier, isStudent: boolean) {
  return tier.prices.find((price) => price.is_student_required === isStudent);
}

export function getFallbackPrice(tier: MembershipTier) {
  return getPriceByStudentStatus(tier, true) ?? tier.prices[0];
}

export function purchaseLabel(tier: Extract<EligibleMembershipTier, { eligible: true }>) {
  if (tier.purchase_type === "upgrade") {
    return `Upgrade to ${tier.title}`;
  }

  return "Choose this pass";
}

const UNAVAILABLE_REASON_LABELS: Record<TierUnavailableReason, string> = {
  already_owned: "Membership already owned",
  not_eligible_current_membership: "Not eligible with current membership",
  executive_restricted: "Executives not eligible",
  competitive_restricted: "Competitive players not eligible",
  purchase_closed: "Purchase opens on {date}",
  unavailable: "Not currently available",
};

const FALLBACK_UNAVAILABLE_MESSAGE = "Not currently available";

// Renders why a tier isn't currently purchasable. Falls back to the generic
// message for a reason the frontend doesn't recognize (eg. after a backend
// deploy adds a new one this build predates), rather than showing nothing
// or throwing.
export function unavailableMessage(tier: Extract<EligibleMembershipTier, { eligible: false }>) {
  const label = UNAVAILABLE_REASON_LABELS[tier.unavailable_reason];
  if (!label) {
    return FALLBACK_UNAVAILABLE_MESSAGE;
  }

  if (tier.unavailable_reason === "purchase_closed") {
    if (!tier.purchase_opens_at) {
      return FALLBACK_UNAVAILABLE_MESSAGE;
    }
    return label.replace(
      "{date}",
      VANCOUVER_DATE_FORMATTER.format(new Date(tier.purchase_opens_at)),
    );
  }

  return label;
}

export function isMembershipTierPrice(
  price: MembershipTierPrice | undefined,
): price is MembershipTierPrice {
  return Boolean(price);
}
