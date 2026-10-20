"use client";

import { ChevronDown, Logs, LogOut, Menu, TicketCheck, UserRound, UsersRound } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useOptionalProfile } from "@/lib/profile.hook";
import { useSignOut } from "@/lib/use-sign-out.hook";
import { cn } from "@/lib/utils/cn";
import { getInitials } from "@/lib/utils/formatting";

const navItems = [
  { href: "/", label: "Home" },
  { href: "/pricing", label: "Pricing" },
];

function DropdownLink({
  href,
  icon,
  children,
}: {
  href: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className="flex items-center gap-3 px-4 py-3 text-sm text-brand-text-muted transition hover:bg-white/5 hover:text-brand-text focus-visible:bg-white/5 focus-visible:text-brand-text focus-visible:outline-none"
    >
      {icon}
      {children}
    </Link>
  );
}

function AdminMenuLinks() {
  return (
    <div className="border-t border-brand-border p-1.5">
      <details className="group/admin">
        <summary className="flex cursor-pointer list-none items-center gap-3 px-2.5 py-2.5 text-sm font-semibold text-brand-text-muted transition hover:bg-white/5 hover:text-brand-text focus-visible:bg-white/5 focus-visible:text-brand-text focus-visible:outline-none [&::-webkit-details-marker]:hidden">
          <UsersRound aria-hidden="true" className="size-4" />
          <span>Admin</span>
          <ChevronDown
            aria-hidden="true"
            className="ml-auto size-3.5 text-brand-text-subtle transition group-open/admin:rotate-180"
          />
        </summary>
        <div className="mt-1 ml-3 border-l border-brand-border">
          <DropdownLink
            href="/admin/users"
            icon={<UsersRound aria-hidden="true" className="size-4" />}
          >
            Users
          </DropdownLink>
          <DropdownLink
            href="/admin/audit-logs"
            icon={<Logs aria-hidden="true" className="size-4" />}
          >
            Audit logs
          </DropdownLink>
          <DropdownLink
            href="/admin/membership-invitations"
            icon={<TicketCheck aria-hidden="true" className="size-4" />}
          >
            Membership Invitations
          </DropdownLink>
        </div>
      </details>
    </div>
  );
}

export function Header() {
  const pathname = usePathname();
  const { data: profile, isPending } = useOptionalProfile();
  const { mutate: signOut, isPending: isSigningOut } = useSignOut();

  return (
    <header className="relative z-30 mx-auto mt-3 w-[calc(100%-1.5rem)] max-w-7xl border border-brand-border/70 bg-brand-surface/90 shadow-2xl shadow-black/25 backdrop-blur-xl md:mt-5 md:w-[calc(100%-2.5rem)]">
      <div className="flex h-16 items-center justify-between px-3 md:hidden">
        <Link
          href="/"
          aria-label="UBCEA Memberships home"
          className="flex items-center gap-2.5 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand-primary"
        >
          <Image
            src="/ubcea_logo.jpg"
            alt=""
            width={36}
            height={36}
            className="size-9 border border-brand-border object-cover"
          />
          <span className="text-sm font-bold tracking-[0.14em] text-brand-text">UBCEA</span>
        </Link>

        <details className="group static">
          <summary
            aria-label="Open navigation menu"
            className="flex h-10 cursor-pointer list-none items-center gap-2 border border-brand-border bg-white/3 px-2.5 text-brand-text transition hover:border-brand-text-muted hover:bg-white/5 focus-visible:outline-2 focus-visible:outline-brand-primary [&::-webkit-details-marker]:hidden"
          >
            {profile?.avatarUrl ? (
              <Image
                src={profile.avatarUrl}
                alt=""
                width={28}
                height={28}
                className="size-7 rounded-full object-cover"
              />
            ) : profile ? (
              <span className="flex size-7 items-center justify-center rounded-full bg-brand-primary/20 text-[0.65rem] font-bold">
                {getInitials(profile.name, profile.email)}
              </span>
            ) : null}
            <Menu aria-hidden="true" className="size-5" />
          </summary>

          <div className="absolute inset-x-0 top-[calc(100%+1px)] w-full border border-brand-border bg-brand-surface shadow-2xl shadow-black/40">
            <nav aria-label="Mobile navigation" className="p-1.5">
              {navItems.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    aria-current={isActive ? "page" : undefined}
                    className={cn(
                      "flex items-center border-l-2 px-3 py-3 text-sm font-semibold transition focus-visible:outline-none",
                      isActive
                        ? "border-brand-primary bg-white/5 text-brand-text"
                        : "border-transparent text-brand-text-muted hover:bg-white/5 hover:text-brand-text",
                    )}
                  >
                    {item.label}
                  </Link>
                );
              })}
            </nav>

            <div aria-hidden="true" className="h-2 border-y border-brand-border bg-brand-bg/60" />

            {isPending ? (
              <div className="m-3 h-10 animate-pulse bg-white/5" />
            ) : profile ? (
              <>
                <div className="border-b border-brand-border bg-white/3 px-4 py-3">
                  <p className="mb-1 text-[0.65rem] font-semibold tracking-[0.14em] text-brand-text-subtle uppercase">
                    Account
                  </p>
                  <p className="truncate text-sm font-semibold text-brand-text">{profile.name}</p>
                  <p className="mt-0.5 truncate text-xs text-brand-text-subtle">{profile.email}</p>
                </div>
                <DropdownLink
                  href="/profile"
                  icon={<UserRound aria-hidden="true" className="size-4" />}
                >
                  My profile
                </DropdownLink>
                {profile.role === "admin" ? <AdminMenuLinks /> : null}
                <div className="border-t border-brand-border p-1.5">
                  <button
                    type="button"
                    disabled={isSigningOut}
                    onClick={() => signOut()}
                    className="flex w-full cursor-pointer items-center gap-3 px-2.5 py-2.5 text-left text-sm text-brand-text-muted transition hover:bg-white/5 hover:text-brand-text focus-visible:bg-white/5 focus-visible:text-brand-text focus-visible:outline-none disabled:cursor-wait disabled:opacity-60"
                  >
                    <LogOut aria-hidden="true" className="size-4" />
                    {isSigningOut ? "Logging out…" : "Log out"}
                  </button>
                </div>
              </>
            ) : (
              <div className="p-3">
                <Link
                  href="/login"
                  className="flex h-11 items-center justify-center bg-brand-primary px-4 text-sm font-semibold text-white transition hover:bg-brand-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-text"
                >
                  Log in
                </Link>
              </div>
            )}
          </div>
        </details>
      </div>

      <div className="hidden h-18 grid-cols-[1fr_auto_1fr] items-center gap-5 px-6 md:grid lg:px-8">
        <Link
          href="/"
          aria-label="UBCEA Memberships home"
          className="flex shrink-0 items-center gap-3 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand-primary"
        >
          <Image
            src="/ubcea_logo.jpg"
            alt=""
            width={38}
            height={38}
            className="size-9.5 border border-brand-border object-cover"
          />
          <span className="hidden leading-tight sm:block">
            <span className="block text-sm font-bold tracking-[0.16em] text-brand-text">
              UBC Esports Portal
            </span>
            <span className="block text-xs text-brand-text-subtle">app.ubcesports.ca</span>
          </span>
        </Link>

        <nav aria-label="Main navigation" className="flex items-center justify-center gap-2">
          {navItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={isActive ? "page" : undefined}
                className={`border-b-2 px-3 py-2 text-sm font-medium transition focus-visible:outline-2 focus-visible:outline-brand-primary ${
                  isActive
                    ? "border-brand-primary text-brand-text"
                    : "border-transparent text-brand-text-muted hover:text-brand-text"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="flex items-center justify-end gap-2">
          {isPending ? (
            <div
              aria-label="Checking login status"
              className="h-10 w-28 animate-pulse border border-brand-border bg-white/5"
            />
          ) : profile ? (
            <details className="group relative">
              <summary className="flex h-11 max-w-52 cursor-pointer list-none items-center gap-2.5 border border-brand-border bg-white/3 px-2 pr-3 transition hover:border-brand-text-muted hover:bg-white/5 focus-visible:outline-2 focus-visible:outline-brand-primary [&::-webkit-details-marker]:hidden">
                {profile.avatarUrl ? (
                  <Image
                    src={profile.avatarUrl}
                    alt=""
                    width={32}
                    height={32}
                    className="size-8 shrink-0 rounded-full object-cover"
                  />
                ) : (
                  <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-brand-primary/20 text-xs font-bold text-brand-text">
                    {getInitials(profile.name, profile.email)}
                  </span>
                )}
                <span className="hidden truncate text-sm font-semibold text-brand-text sm:block">
                  {profile.name}
                </span>
                <ChevronDown
                  aria-hidden="true"
                  className="size-3.5 shrink-0 text-brand-text-subtle transition group-open:rotate-180"
                />
              </summary>

              <div className="absolute right-0 top-[calc(100%+0.5rem)] w-64 border border-brand-border bg-brand-surface shadow-2xl shadow-black/35">
                <div className="border-b border-brand-border px-4 py-3">
                  <p className="truncate text-sm font-semibold text-brand-text">{profile.name}</p>
                  <p className="mt-0.5 truncate text-xs text-brand-text-subtle">{profile.email}</p>
                </div>
                <DropdownLink
                  href="/profile"
                  icon={<UserRound aria-hidden="true" className="size-4" />}
                >
                  My profile
                </DropdownLink>
                {profile.role === "admin" ? <AdminMenuLinks /> : null}
                <div className="border-t border-brand-border p-1.5">
                  <button
                    type="button"
                    disabled={isSigningOut}
                    onClick={() => signOut()}
                    className="flex w-full cursor-pointer items-center gap-3 px-2.5 py-2 text-left text-sm text-brand-text-muted transition hover:bg-white/5 hover:text-brand-text focus-visible:bg-white/5 focus-visible:text-brand-text focus-visible:outline-none disabled:cursor-wait disabled:opacity-60"
                  >
                    <LogOut aria-hidden="true" className="size-4" />
                    {isSigningOut ? "Logging out…" : "Log out"}
                  </button>
                </div>
              </div>
            </details>
          ) : (
            <Link
              href="/login"
              className="inline-flex h-10 items-center justify-center border border-brand-primary bg-brand-primary px-4 text-sm font-semibold text-white transition hover:border-brand-primary-hover hover:bg-brand-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-text"
            >
              Log in
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
