import { Check, ShieldCheck, X } from "lucide-react";
import { PurchaseButton } from "@/components/membership/purchase-button";
import { formatMembershipPrice } from "@/components/membership/pricing";
import type { EligibleMembershipTier } from "@/lib/types/membership.types";

type AssignedPassCardProps = {
  tier: Extract<EligibleMembershipTier, { eligible: true }>;
  checkoutPending: boolean;
  onCheckout: (tier: Extract<EligibleMembershipTier, { eligible: true }>) => void;
};

export function AssignedPassCard({ tier, checkoutPending, onCheckout }: AssignedPassCardProps) {
  return (
    <article className="grid border border-brand-primary/40 bg-brand-primary/10 lg:grid-cols-[minmax(0,1fr)_minmax(19rem,0.55fr)]">
      <div className="flex gap-5 p-6 sm:p-7">
        <div className="flex size-11 shrink-0 items-center justify-center border border-brand-primary/40 bg-brand-primary/15">
          <ShieldCheck aria-hidden="true" className="size-5 text-blue-100" />
        </div>
        <div>
          <p className="mb-2 font-mono text-xs font-semibold uppercase tracking-[0.2em] text-blue-100">
            Assigned pass
          </p>
          <h3 className="text-xl font-semibold text-brand-text">{tier.title}</h3>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-brand-text-muted">
            {tier.description}
          </p>
          {(tier.benefits ?? []).length > 0 || (tier.limitations ?? []).length > 0 ? (
            <ul className="mt-4 grid gap-2 text-sm text-brand-text-muted">
              {(tier.benefits ?? []).map((benefit) => (
                <li key={benefit} className="flex gap-3">
                  <Check aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-blue-200" />
                  <span>{benefit}</span>
                </li>
              ))}
              {(tier.limitations ?? []).map((limitation) => (
                <li key={limitation} className="flex gap-3">
                  <X aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-red-300" />
                  <span>{limitation}</span>
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      </div>

      <div className="border-t border-brand-primary/30 p-6 lg:border-l lg:border-t-0">
        <div className="mb-4 flex items-end justify-between gap-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-wider text-brand-text-subtle">
              Your assigned price
            </p>
            <p className="mt-1 text-3xl font-semibold text-brand-text">
              {formatMembershipPrice(tier.prices.price)}
            </p>
          </div>
          <Check aria-hidden="true" className="mb-2 size-4 text-blue-200" />
        </div>
        <PurchaseButton
          tier={tier}
          isSignedIn
          needsOnboarding={false}
          checkoutPending={checkoutPending}
          onCheckout={onCheckout}
          onSignIn={() => undefined}
          signInPending={false}
          featured
        />
      </div>
    </article>
  );
}
