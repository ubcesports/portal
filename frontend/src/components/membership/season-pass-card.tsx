import { CalendarDays, Check, Sparkles, X } from "lucide-react";
import { PurchaseButton } from "@/components/membership/purchase-button";
import {
  formatMembershipExpiration,
  formatMembershipPrice,
  getPriceByStudentStatus,
  isMembershipTierPrice,
  membershipPriceLabel,
} from "@/components/membership/pricing";
import type { EligibleMembershipTier, MembershipTier } from "@/lib/types/membership.types";

type SeasonPassCardProps = {
  tier: MembershipTier;
  eligibleTier?: EligibleMembershipTier;
  checkoutPending: boolean;
  isSignedIn: boolean;
  needsOnboarding: boolean;
  onCheckout: (tier: EligibleMembershipTier) => void;
  onSignIn: () => void;
  signInPending: boolean;
};

export function SeasonPassCard({
  tier,
  eligibleTier,
  checkoutPending,
  isSignedIn,
  needsOnboarding,
  onCheckout,
  onSignIn,
  signInPending,
}: SeasonPassCardProps) {
  const featured = tier.slug === "lounge";

  return (
    <article
      className={`relative flex min-h-124 flex-col border bg-brand-surface/88 shadow-2xl shadow-black/20 ${
        featured ? "border-brand-primary shadow-brand-primary/10" : "border-brand-border"
      }`}
    >
      {featured ? <div className="absolute inset-x-0 top-0 h-1 bg-brand-primary" /> : null}
      <div className="flex items-start justify-between gap-4 border-b border-brand-border px-6 py-6 sm:px-7">
        <div>
          <p className="mb-2 font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-text-subtle">
            Season pass
          </p>
          <h2 className="text-2xl font-semibold text-brand-text">{tier.title}</h2>
        </div>
        {featured ? (
          <span className="inline-flex items-center gap-1.5 border border-brand-primary/70 bg-brand-primary/15 px-3 py-1.5 text-xs font-semibold text-blue-100">
            <Sparkles aria-hidden="true" className="size-3.5" />
            Popular
          </span>
        ) : null}
      </div>

      <div className="flex flex-1 flex-col p-6 sm:p-7">
        {eligibleTier ? (
          <div className="border border-brand-primary/35 bg-brand-primary/10 p-5">
            <div className="flex items-center justify-between gap-3">
              <span className="text-sm font-medium text-blue-100">
                {eligibleTier.purchase_type === "upgrade"
                  ? "Your upgrade price"
                  : "Your eligible price"}
              </span>
              <Check aria-hidden="true" className="size-4 text-blue-200" />
            </div>
            <div className="mt-3 flex items-end gap-2">
              <span className="text-4xl font-semibold tracking-tight text-brand-text">
                {formatMembershipPrice(eligibleTier.prices.price)}
              </span>
              <span className="pb-1 text-sm text-brand-text-subtle">one time</span>
            </div>
          </div>
        ) : (
          <PublicPriceGrid tier={tier} />
        )}

        <div className="mt-6 flex flex-1 flex-col">
          <div>
            <p className="text-sm leading-6 text-brand-text-muted">
              {tier.description || `${tier.title} UBCEA membership pass.`}
            </p>
            <BenefitList limitations={tier.limitations} benefits={tier.benefits} />
          </div>

          <div className="mt-auto pt-5">
            <div className="flex items-start gap-3 border-t border-brand-border/70 pt-5">
              <CalendarDays
                aria-hidden="true"
                className="mt-0.5 size-4 shrink-0 text-brand-text-subtle"
              />
              <div>
                <p className="text-sm font-medium leading-6 text-brand-text">
                  {formatMembershipExpiration(tier.expiration_type)}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="mt-7">
          <PurchaseButton
            tier={eligibleTier}
            isSignedIn={isSignedIn}
            needsOnboarding={needsOnboarding}
            checkoutPending={checkoutPending}
            onCheckout={onCheckout}
            onSignIn={onSignIn}
            signInPending={signInPending}
            featured={featured}
          />
        </div>
      </div>
    </article>
  );
}

function BenefitList({ benefits, limitations }: { benefits: string[]; limitations: string[] }) {
  if (benefits.length === 0 && limitations.length === 0) {
    return null;
  }

  return (
    <ul className="mt-5 grid gap-2 border-t border-brand-border/70 pt-5 text-sm text-brand-text-muted">
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

function PublicPriceGrid({ tier }: { tier: MembershipTier }) {
  const studentPrice = getPriceByStudentStatus(tier, true);
  const communityPrice = getPriceByStudentStatus(tier, false);
  const prices = [studentPrice, communityPrice].filter(isMembershipTierPrice);

  return (
    <div className="grid grid-cols-2 border border-brand-border">
      {prices.map((price, index) => (
        <div
          key={price.price_id}
          className={`p-4 sm:p-5 ${index === 0 ? "border-r border-brand-border" : ""}`}
        >
          <p className="text-xs font-medium uppercase tracking-wider text-brand-text-subtle">
            {membershipPriceLabel(price)}
          </p>
          <p className="mt-2 text-2xl font-semibold text-brand-text">
            {formatMembershipPrice(price.price)}
          </p>
        </div>
      ))}
    </div>
  );
}
