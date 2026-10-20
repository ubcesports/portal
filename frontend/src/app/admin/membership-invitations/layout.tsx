import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Membership Invitations",
};

export default function MembershipInvitationsLayout({ children }: { children: ReactNode }) {
  return children;
}
