export type AnnouncementConfig = {
  enabled: boolean;
  message: string;
  linkLabel: string;
  href: string;
};

export const announcement: AnnouncementConfig = {
  enabled: true,
  message: "Liftoff attendees get $5 off all memberships",
  linkLabel: "View event",
  href: "https://ubcesports.ca/events",
};
