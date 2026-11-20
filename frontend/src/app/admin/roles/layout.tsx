import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Roles",
};

export default function AdminRolesLayout({ children }: { children: ReactNode }) {
  return children;
}
