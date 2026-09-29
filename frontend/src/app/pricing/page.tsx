import type { Metadata } from "next";
import type { MembershipTier } from "@/lib/types/membership.types";
import { SITE_URL } from "@/lib/site";
import { PricingClient } from "./pricing-client";

const PRICING_DESCRIPTION =
  "Compare UBC Esports Association membership passes, public pricing, eligibility, and benefits, including Legion Gaming Lounge access.";

const RESTRICTED_TIER_SLUGS = new Set(["competitive_team", "executive"]);

export const metadata: Metadata = {
  title: "Membership Pricing",
  description: PRICING_DESCRIPTION,
  alternates: {
    canonical: "/pricing",
  },
  openGraph: {
    url: "/pricing",
    title: "UBC Esports Membership Pricing",
    description: PRICING_DESCRIPTION,
    images: [
      {
        url: "/opengraph-image",
        width: 1200,
        height: 630,
        alt: "UBC Esports Memberships",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "UBC Esports Membership Pricing",
    description: PRICING_DESCRIPTION,
    images: ["/opengraph-image"],
  },
};

export const dynamic = "force-dynamic";

const API_BASE =
  (process.env.API_INTERNAL_URL || process.env.NEXT_PUBLIC_API_URL)?.replace(/\/$/, "") ||
  "http://localhost:8080";

async function getMembershipCatalog(): Promise<MembershipTier[] | undefined> {
  try {
    const response = await fetch(`${API_BASE}/membership/tiers`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return undefined;
    }

    return (await response.json()) as MembershipTier[];
  } catch {
    // The client query keeps the page usable if the API is temporarily unavailable to the server.
    return undefined;
  }
}

function getCatalogJsonLd(catalog: MembershipTier[]) {
  const publicCatalog = catalog.filter((tier) => !RESTRICTED_TIER_SLUGS.has(tier.slug));

  return {
    "@context": "https://schema.org",
    "@type": "ItemList",
    name: "UBC Esports membership passes",
    url: `${SITE_URL}/pricing`,
    itemListElement: publicCatalog.map((tier, index) => ({
      "@type": "ListItem",
      position: index + 1,
      item: {
        "@type": "Product",
        name: tier.title,
        description: tier.description || `${tier.title} UBC Esports membership pass.`,
        category: "Membership",
        brand: {
          "@type": "Organization",
          name: "UBC Esports Association",
        },
        offers: tier.prices.map((price) => ({
          "@type": "Offer",
          priceCurrency: "CAD",
          price: price.price.toFixed(2),
          url: `${SITE_URL}/pricing`,
        })),
      },
    })),
  };
}

export default async function PricingPage() {
  const catalog = await getMembershipCatalog();
  const catalogJsonLd = catalog ? getCatalogJsonLd(catalog) : undefined;

  return (
    <>
      {catalogJsonLd ? (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(catalogJsonLd).replace(/</g, "\\u003c"),
          }}
        />
      ) : null}
      <PricingClient initialCatalog={catalog} />
    </>
  );
}
