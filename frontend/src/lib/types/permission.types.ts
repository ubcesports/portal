/*
  Permission keys granted through roles. Keep in sync with
  backend/internal/auth/permissions.go and the permissions migration.

  Shared by the role management pages and the permission-aware admin frontend,
  so add new keys here rather than redeclaring them elsewhere.
*/
export const PERMISSIONS = {
  UsersView: "users.view",
  UsersUpdate: "users.update",
  UsersExport: "users.export",

  MembershipTiersView: "membership_tiers.view",

  MembershipsView: "memberships.view",
  MembershipsCreate: "memberships.create",
  MembershipsDelete: "memberships.delete",

  MembershipInvitationsView: "membership_invitations.view",
  MembershipInvitationsCreate: "membership_invitations.create",
  MembershipInvitationsDelete: "membership_invitations.delete",

  ExecProfilesUpdate: "exec_profiles.update",

  AuditLogsView: "audit_logs.view",
  AuditLogsExport: "audit_logs.export",

  AnalyticsView: "analytics.view",
  AnalyticsExport: "analytics.export",

  LoungeSessionsCreate: "lounge_sessions.create",
  LoungeSessionsCheckout: "lounge_sessions.checkout",
  LoungeSessionsView: "lounge_sessions.view",
  LoungeStationsUpdate: "lounge_stations.update",
} as const;

export type PermissionKey = (typeof PERMISSIONS)[keyof typeof PERMISSIONS];

export const ALL_PERMISSION_KEYS: PermissionKey[] = Object.values(PERMISSIONS);

export type PermissionOption = {
  key: PermissionKey;
  label: string;
  description: string;
};

export type PermissionGroup = {
  label: string;
  permissions: PermissionOption[];
};

/* Permissions grouped by area, in the order the admin UI lists them. */
export const PERMISSION_GROUPS: PermissionGroup[] = [
  {
    label: "Users",
    permissions: [
      { key: PERMISSIONS.UsersView, label: "View users", description: "Browse and open users." },
      {
        key: PERMISSIONS.UsersUpdate,
        label: "Edit users",
        description: "Change names, student details and groups.",
      },
      {
        key: PERMISSIONS.UsersExport,
        label: "Export users",
        description: "Download users as CSV.",
      },
    ],
  },
  {
    label: "Memberships",
    permissions: [
      {
        key: PERMISSIONS.MembershipTiersView,
        label: "View membership tiers",
        description: "See the tier list used in filters and forms.",
      },
      {
        key: PERMISSIONS.MembershipsView,
        label: "View memberships",
        description: "See a user's membership history.",
      },
      {
        key: PERMISSIONS.MembershipsCreate,
        label: "Add memberships",
        description: "Record offline membership purchases.",
      },
      {
        key: PERMISSIONS.MembershipsDelete,
        label: "Cancel memberships",
        description: "Cancel a user's membership.",
      },
    ],
  },
  {
    label: "Membership invitations",
    permissions: [
      {
        key: PERMISSIONS.MembershipInvitationsView,
        label: "View invitations",
        description: "See pending and completed invitations.",
      },
      {
        key: PERMISSIONS.MembershipInvitationsCreate,
        label: "Create invitations",
        description: "Record in-person purchases and email invites.",
      },
      {
        key: PERMISSIONS.MembershipInvitationsDelete,
        label: "Delete invitations",
        description: "Remove pending invitations.",
      },
    ],
  },
  {
    label: "Exec profiles",
    permissions: [
      {
        key: PERMISSIONS.ExecProfilesUpdate,
        label: "Edit exec profiles",
        description: "Change exec titles, groups and ordering.",
      },
    ],
  },
  {
    label: "Audit logs",
    permissions: [
      {
        key: PERMISSIONS.AuditLogsView,
        label: "View audit logs",
        description: "Browse admin actions.",
      },
      {
        key: PERMISSIONS.AuditLogsExport,
        label: "Export audit logs",
        description: "Download audit logs as CSV.",
      },
    ],
  },
  {
    label: "Analytics",
    permissions: [
      {
        key: PERMISSIONS.AnalyticsView,
        label: "View analytics",
        description: "See membership and revenue charts.",
      },
      {
        key: PERMISSIONS.AnalyticsExport,
        label: "Export analytics",
        description: "Download analytics as CSV.",
      },
    ],
  },
  {
    label: "Lounge",
    permissions: [
      {
        key: PERMISSIONS.LoungeSessionsView,
        label: "View lounge sessions",
        description: "See lounge session history.",
      },
      {
        key: PERMISSIONS.LoungeSessionsCreate,
        label: "Check members in",
        description: "Start lounge sessions.",
      },
      {
        key: PERMISSIONS.LoungeSessionsCheckout,
        label: "Check members out",
        description: "End lounge sessions.",
      },
      {
        key: PERMISSIONS.LoungeStationsUpdate,
        label: "Manage stations",
        description: "Change lounge station details.",
      },
    ],
  },
];

const PERMISSION_LABELS = new Map<string, string>(
  PERMISSION_GROUPS.flatMap((group) =>
    group.permissions.map((permission) => [permission.key, permission.label] as const),
  ),
);

/* Returns the readable label for a key, or the raw key if it is unknown. */
export function getPermissionLabel(key: string): string {
  return PERMISSION_LABELS.get(key) ?? key;
}
