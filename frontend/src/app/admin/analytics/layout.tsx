import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Analytics",
};

export default function AdminAnalyticsLayout({ children }: { children: ReactNode }) {
  return children;
}
