import { ArrowRight, Loader2, LogIn } from "lucide-react";
import Link from "next/link";
import { purchaseLabel, unavailableMessage } from "@/components/membership/pricing";
import type { EligibleMembershipTier } from "@/lib/types/membership.types";

type PurchaseButtonProps = {
  tier?: EligibleMembershipTier;
  isSignedIn: boolean;
  needsOnboarding: boolean;
  checkoutPending: boolean;
  onCheckout: (tier: Extract<EligibleMembershipTier, { eligible: true }>) => void;
  onSignIn: () => void;
  signInPending: boolean;
  featured?: boolean;
};

export function PurchaseButton({
  tier,
  isSignedIn,
  needsOnboarding,
  checkoutPending,
  onCheckout,
  onSignIn,
  signInPending,
  featured = false,
}: PurchaseButtonProps) {
  const baseClass = featured
    ? "border-brand-primary bg-brand-primary hover:bg-brand-primary-hover"
    : "border-brand-border bg-white/[0.04] hover:border-brand-text-muted hover:bg-white/[0.08]";

  if (tier?.eligible) {
    return (
      <button
        type="button"
        onClick={() => onCheckout(tier)}
        disabled={checkoutPending}
        className={`inline-flex h-12 w-full cursor-pointer items-center justify-center gap-2 border px-5 text-sm font-semibold text-brand-text transition disabled:cursor-not-allowed disabled:opacity-60 ${baseClass}`}
      >
        {checkoutPending ? (
          <Loader2 aria-hidden="true" className="size-4 animate-spin" />
        ) : (
          <ArrowRight aria-hidden="true" className="size-4" />
        )}
        {checkoutPending ? "Opening checkout" : purchaseLabel(tier)}
      </button>
    );
  }

  if (!isSignedIn) {
    return (
      <button
        type="button"
        onClick={onSignIn}
        disabled={signInPending}
        className={`inline-flex h-12 w-full cursor-pointer items-center justify-center gap-2 border px-5 text-sm font-semibold text-brand-text transition disabled:cursor-not-allowed disabled:opacity-60 ${baseClass}`}
      >
        {signInPending ? (
          <Loader2 aria-hidden="true" className="size-4 animate-spin" />
        ) : (
          <LogIn aria-hidden="true" className="size-4" />
        )}
        {signInPending ? "Connecting" : "Sign in to purchase"}
      </button>
    );
  }

  if (needsOnboarding) {
    return (
      <Link
        href="/onboard"
        className={`inline-flex h-12 w-full items-center justify-center gap-2 border px-5 text-sm font-semibold text-brand-text transition ${baseClass}`}
      >
        <ArrowRight aria-hidden="true" className="size-4" />
        Finish account setup
      </Link>
    );
  }

  return (
    <div className="flex h-12 w-full items-center justify-center border border-brand-border bg-white/2 px-5 text-center text-sm font-medium text-brand-text-muted">
      {tier ? unavailableMessage(tier) : "Not currently available"}
    </div>
  );
}
