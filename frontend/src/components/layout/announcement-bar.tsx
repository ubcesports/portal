import { ArrowUpRight, TicketPercent } from "lucide-react";
import { announcement } from "@/lib/announcement";

export function AnnouncementBar() {
  if (!announcement.enabled) {
    return null;
  }

  return (
    <aside aria-label="Announcement" className="relative z-40 bg-brand-primary text-white">
      <a
        href={announcement.href}
        className="group mx-auto flex min-h-10 w-full max-w-7xl items-center justify-center gap-2 px-5 py-2 text-center text-xs font-semibold tracking-[0.01em] transition hover:bg-white/10 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-white sm:text-sm"
      >
        <TicketPercent aria-hidden="true" className="size-4 shrink-0" />
        <span>{announcement.message}</span>
        <span className="hidden items-center gap-1 whitespace-nowrap underline decoration-white/60 underline-offset-4 sm:inline-flex">
          {announcement.linkLabel}
          <ArrowUpRight
            aria-hidden="true"
            className="size-3.5 transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
          />
        </span>
        <ArrowUpRight aria-hidden="true" className="size-3.5 shrink-0 sm:hidden" />
      </a>
    </aside>
  );
}
