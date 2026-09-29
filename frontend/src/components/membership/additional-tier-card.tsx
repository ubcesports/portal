import { CalendarDays, Check, Ticket, X } from "lucide-react";
import { PurchaseButton } from "@/components/membership/purchase-button";
import {
  formatMembershipExpiration,
  formatMembershipPrice,
  membershipPriceLabel,
} from "@/components/membership/pricing";
import type { EligibleMembershipTier, MembershipTier } from "@/lib/types/membership.types";

type AdditionalTierCardProps = {
  tier: MembershipTier;
  eligibleTier?: EligibleMembershipTier;
  checkoutPending: boolean;
  isSignedIn: boolean;
  needsOnboarding: boolean;
  onCheckout: (tier: Extract<EligibleMembershipTier, { eligible: true }>) => void;
  onSignIn: () => void;
  signInPending: boolean;
};

export function AdditionalTierCard({
  tier,
  eligibleTier,
  checkoutPending,
  isSignedIn,
  needsOnboarding,
  onCheckout,
  onSignIn,
  signInPending,
}: AdditionalTierCardProps) {
  return (
    <article className="grid border border-brand-border bg-brand-surface/75 lg:grid-cols-[minmax(0,1fr)_minmax(19rem,0.65fr)]">
      <div className="flex gap-5 p-6 sm:p-7">
        <div className="flex size-11 shrink-0 items-center justify-center border border-brand-border bg-white/4">
          <Ticket aria-hidden="true" className="size-5 text-brand-text-muted" />
        </div>
        <div>
          <p className="mb-2 font-mono text-xs font-semibold uppercase tracking-[0.18em] text-brand-text-subtle">
            {tier.program_name}
          </p>
          <h3 className="text-xl font-semibold text-brand-text">{tier.title}</h3>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-brand-text-muted">
            {tier.description || `${tier.title} UBCEA membership pass.`}
          </p>
          <BenefitList limitations={tier.limitations} benefits={tier.benefits} />
          {!eligibleTier?.eligible ? <InlinePublicPrices tier={tier} /> : null}
        </div>
      </div>

      <div className="border-t border-brand-border p-6 lg:border-l lg:border-t-0">
        {eligibleTier?.eligible ? (
          <div className="mb-4 flex items-end justify-between gap-4">
            <div>
              <p className="text-xs font-medium uppercase tracking-wider text-brand-text-subtle">
                {eligibleTier.purchase_type === "upgrade"
                  ? "Your upgrade price"
                  : "Your eligible price"}
              </p>
              <p className="mt-1 text-3xl font-semibold text-brand-text">
                {formatMembershipPrice(eligibleTier.prices.price)}
              </p>
            </div>
            <Check aria-hidden="true" className="mb-2 size-4 text-blue-200" />
          </div>
        ) : null}
        <div className="mb-5 flex items-start gap-3 border-t border-brand-border/70 pt-4">
          <CalendarDays
            aria-hidden="true"
            className="mt-0.5 size-4 shrink-0 text-brand-text-subtle"
          />
          <p className="text-sm font-medium leading-6 text-brand-text">
            {formatMembershipExpiration(tier.expiration_type)}
          </p>
        </div>
        <PurchaseButton
          tier={eligibleTier}
          isSignedIn={isSignedIn}
          needsOnboarding={needsOnboarding}
          checkoutPending={checkoutPending}
          onCheckout={onCheckout}
          onSignIn={onSignIn}
          signInPending={signInPending}
        />
      </div>
    </article>
  );
}

function BenefitList({
  benefits = [],
  limitations = [],
}: {
  benefits?: string[];
  limitations?: string[];
}) {
  if (benefits.length === 0 && limitations.length === 0) {
    return null;
  }

  return (
    <ul className="mt-4 grid gap-2 text-sm text-brand-text-muted">
      {benefits.map((benefit) => (
        <li key={benefit} className="flex gap-3">
          <Check aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-blue-200" />
          <span>{benefit}</span>
        </li>
      ))}
      {limitations.map((limitation) => (
        <li key={limitation} className="flex gap-3">
          <X aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-red-300" />
          <span>{limitation}</span>
        </li>
      ))}
    </ul>
  );
}

function InlinePublicPrices({ tier }: { tier: MembershipTier }) {
  return (
    <div className="mt-4 flex flex-wrap gap-x-6 gap-y-2 text-sm">
      {tier.prices.map((price) => (
        <span key={price.price_id} className="text-brand-text-muted">
          {membershipPriceLabel(price)}{" "}
          <strong className="font-semibold text-brand-text">
            {formatMembershipPrice(price.price)}
          </strong>
        </span>
      ))}
    </div>
  );
}
