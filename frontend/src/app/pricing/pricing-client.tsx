"use client";

import { Check } from "lucide-react";
import { useMutation } from "@tanstack/react-query";
import { BasePage } from "@/components/layout/base-page";
import { AdditionalTierCard } from "@/components/membership/additional-tier-card";
import { AssignedPassCard } from "@/components/membership/assigned-pass-card";
import { SeasonPassCard } from "@/components/membership/season-pass-card";
import { redirectToSignIn } from "@/lib/auth";
import apiClient from "@/lib/client";
import type { EligibleMembershipTier, MembershipTier } from "@/lib/types/membership.types";
import { useEligibleMembershipTiers, useMembershipCatalog } from "@/lib/membership.hook";
import { useOptionalProfile } from "@/lib/profile.hook";
import { notNull } from "@/lib/utils/type-guards";

import type { CheckoutResponse, CheckoutRequest } from "@/lib/types/membership.types";

const RESTRICTED_TIER_SLUGS = ["competitive_team", "executive"];
const MAIN_TIER_SLUGS = ["basic", "lounge"];

type PricingClientProps = {
  initialCatalog?: MembershipTier[];
};

export function PricingClient({ initialCatalog }: PricingClientProps) {
  const {
    data: catalog,
    isPending: catalogPending,
    isError: catalogError,
  } = useMembershipCatalog(initialCatalog);
  const { data: profile } = useOptionalProfile();
  const isSignedIn = !!profile;
  const needsOnboarding = isSignedIn && !profile.onboardingCompletedAt;
  const canLoadEligibility = !!profile?.onboardingCompletedAt;
  const {
    data: eligibleTiers = [],
    isPending: eligibilityPending,
    isError: eligibilityError,
  } = useEligibleMembershipTiers(canLoadEligibility);

  const { mutate: signIn, isPending: signInPending } = useMutation({
    mutationFn: async () => await redirectToSignIn(window.location.href),
  });

  const {
    mutate: checkout,
    variables: checkoutTier,
    isPending: checkoutPending,
  } = useMutation({
    mutationFn: async (tier: EligibleMembershipTier) => {
      const response = await apiClient.post<CheckoutResponse>("/membership/checkout", {
        tier_id: tier.id,
      } satisfies CheckoutRequest);

      window.location.assign(response.data.url);
    },
  });

  const tierBySlug = (slug: string) => catalog?.find((tier) => tier.slug === slug);
  const eligibleById = (tierId: string) => eligibleTiers.find((tier) => tier.id === tierId);

  const mainTiers = MAIN_TIER_SLUGS.map(tierBySlug).filter(notNull);
  const assignedTiers = eligibleTiers.filter((tier) => RESTRICTED_TIER_SLUGS.includes(tier.slug));
  const additionalTiers =
    catalog?.filter(
      (tier) => !MAIN_TIER_SLUGS.includes(tier.slug) && !RESTRICTED_TIER_SLUGS.includes(tier.slug),
    ) ?? [];

  return (
    <BasePage>
      <section className="pb-12 pt-16 text-center sm:pb-14 sm:pt-20">
        <p className="font-mono text-xs font-semibold uppercase tracking-[0.24em] text-blue-200">
          UBCEA membership passes
        </p>
        <h1 className="mx-auto mt-5 max-w-3xl text-4xl font-semibold text-brand-text sm:text-5xl lg:text-6xl">
          Choose your UBCEA pass
        </h1>
        <p className="mx-auto mt-5 max-w-2xl text-base leading-7 text-brand-text-muted sm:text-lg">
          Compare Basic and Lounge tier membership pricing, then sign in to see the lowest price
          available to your account.
        </p>
      </section>

      {profile && !profile.onboardingCompletedAt ? (
        <div className="mb-6 border border-amber-300/35 bg-amber-300/10 px-5 py-4 text-sm text-amber-100">
          Finish your account setup to see your eligible membership prices and purchase a pass.
        </div>
      ) : profile && eligibilityError ? (
        <div className="mb-6 border border-red-400/35 bg-red-400/10 px-5 py-4 text-sm text-red-100">
          Your personalized prices could not be loaded. Refresh the page to try again.
        </div>
      ) : profile ? (
        <div className="mb-6 flex flex-col gap-2 border border-brand-primary/35 bg-brand-primary/10 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <Check aria-hidden="true" className="size-4 shrink-0 text-blue-200" />
            <p className="text-sm text-blue-50">
              Showing eligible checkout options for {profile.name || profile.email}.
            </p>
          </div>
          <p className="text-xs font-medium text-blue-100">
            Restricted passes stay hidden unless assigned to your account.
          </p>
        </div>
      ) : null}

      {catalogPending || (canLoadEligibility && eligibilityPending) ? (
        <div className="grid gap-5 lg:grid-cols-2" aria-label="Loading membership passes">
          {[0, 1].map((item) => (
            <div
              key={item}
              className="h-124 animate-pulse border border-brand-border bg-brand-surface/60"
            />
          ))}
        </div>
      ) : catalogError || !catalog ? (
        <div className="border border-red-400/35 bg-red-400/10 px-6 py-10 text-center text-sm text-red-100">
          Membership prices are unavailable right now. Refresh the page to try again.
        </div>
      ) : (
        <>
          {assignedTiers.length > 0 ? (
            <section className="pb-10" aria-labelledby="assigned-passes-heading">
              <div className="mb-5">
                <p className="font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-text-subtle">
                  Assigned access
                </p>
                <h2
                  id="assigned-passes-heading"
                  className="mt-2 text-xl font-semibold text-brand-text"
                >
                  Your restricted passes
                </h2>
              </div>
              <div className="grid gap-5">
                {assignedTiers.map((tier) => (
                  <AssignedPassCard
                    key={tier.id}
                    tier={tier}
                    checkoutPending={checkoutPending && checkoutTier?.id === tier.id}
                    onCheckout={checkout}
                  />
                ))}
              </div>
            </section>
          ) : null}

          <section aria-labelledby="season-passes-heading">
            <div className="mb-5 flex items-end justify-between gap-4">
              <div>
                <p className="font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-text-subtle">
                  Main passes
                </p>
                <h2
                  id="season-passes-heading"
                  className="mt-2 text-xl font-semibold text-brand-text"
                >
                  Season memberships
                </h2>
              </div>
              <p className="hidden text-sm text-brand-text-subtle sm:block">One-time payment</p>
            </div>
            <div className="grid gap-5 lg:grid-cols-2">
              {mainTiers.map((tier) => (
                <SeasonPassCard
                  key={tier.id}
                  tier={tier}
                  eligibleTier={eligibleById(tier.id)}
                  checkoutPending={checkoutPending && checkoutTier?.id === tier.id}
                  isSignedIn={isSignedIn}
                  needsOnboarding={needsOnboarding}
                  onCheckout={checkout}
                  onSignIn={() => signIn()}
                  signInPending={signInPending}
                />
              ))}
            </div>
          </section>

          {additionalTiers.length > 0 ? (
            <section className="pb-20 pt-12" aria-labelledby="additional-passes-heading">
              <div className="mb-5">
                <p className="font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-text-subtle">
                  More ways to join
                </p>
                <h2
                  id="additional-passes-heading"
                  className="mt-2 text-xl font-semibold text-brand-text"
                >
                  Additional memberships
                </h2>
              </div>
              <div className="grid gap-5">
                {additionalTiers.map((tier) => (
                  <AdditionalTierCard
                    key={tier.id}
                    tier={tier}
                    eligibleTier={eligibleById(tier.id)}
                    checkoutPending={checkoutPending && checkoutTier?.id === tier.id}
                    isSignedIn={isSignedIn}
                    needsOnboarding={needsOnboarding}
                    onCheckout={checkout}
                    onSignIn={() => signIn()}
                    signInPending={signInPending}
                  />
                ))}
              </div>
            </section>
          ) : null}
        </>
      )}
    </BasePage>
  );
}
