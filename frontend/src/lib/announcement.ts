export type AnnouncementConfig = {
  enabled: boolean;
  message: string;
  linkLabel: string;
  href: string;
  expiresAt: string | null;
};

export const announcement: AnnouncementConfig = {
  enabled: true,
  message: "Liftoff attendees get $5 off all memberships",
  linkLabel: "View event",
  href: "https://ubcesports.ca/events",
  expiresAt: "2026-11-19T16:20:00-07:00",
};

