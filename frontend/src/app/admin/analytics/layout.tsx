import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Analytics",
};

/** Renders analytics children beneath the route metadata without an extra wrapper. */
export default function AdminAnalyticsLayout({ children }: { children: ReactNode }) {
  return children;
}
