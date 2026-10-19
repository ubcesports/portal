import type { ReactNode } from "react";
import { AnnouncementBar } from "@/components/layout/announcement-bar";
import { Footer } from "@/components/layout/footer";
import { Header } from "@/components/layout/header";

type BasePageProps = {
  children: ReactNode;
  className?: string;
};

export function BasePage({ children, className = "" }: BasePageProps) {
  return (
    <div
      className={`relative isolate min-h-screen overflow-hidden bg-brand-bg text-brand-text ${className}`}
    >
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 z-0 bg-size-[72px_72px]"
        style={{
          backgroundImage:
            "linear-gradient(to right, color-mix(in srgb, var(--color-border) 18%, transparent) 1px, transparent 1px), linear-gradient(to bottom, color-mix(in srgb, var(--color-border) 14%, transparent) 1px, transparent 1px)",
        }}
      />
      <div className="relative z-10 flex min-h-screen flex-col">
        <AnnouncementBar />
        <Header />
        <main className="relative isolate flex flex-1 overflow-hidden">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-56 border-b border-brand-border/50 bg-brand-surface/45"
          />
          <div
            aria-hidden="true"
            className="pointer-events-none absolute -right-72 top-20 -z-10 h-72 w-208 rotate-[-18deg] border-y border-brand-primary/20 bg-brand-primary/10"
          />
          <div
            aria-hidden="true"
            className="pointer-events-none absolute bottom-0 left-0 -z-10 h-44 w-full border-t border-brand-border/40 bg-brand-surface/30"
          />

          <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col px-5 py-6 sm:px-8 lg:px-10">
            {children}
          </div>
        </main>
      </div>
      <div className="relative z-10">
        <Footer />
      </div>
    </div>
  );
}
