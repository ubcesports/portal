import apiClient from "../client";
import type {
  AdminUserFilters,
  AdminMembershipTierOption,
  AdminPagination,
  AppliedSearch,
  UpdateUserRequest,
  UserResponse,
  UsersResponse,
  AuditLogResponse,
  AddOfflineMembershipRequest,
  CreateMembershipInvitationRequest,
  MembershipInvitation,
} from "@/lib/types/admin.types";
import type { User } from "@/lib/types/user.types";
import type { EligibleMembershipTier, Membership } from "../types/membership.types";

export function buildAdminUserParams(
  appliedSearch: AppliedSearch,
  filters: AdminUserFilters,
  pagination?: AdminPagination,
): URLSearchParams {
  const params = new URLSearchParams();

  if (appliedSearch?.value.trim()) {
    params.set(appliedSearch.mode, appliedSearch.value.trim());
  }

  if (filters.role) {
    params.set("role", filters.role);
  }

  for (const group of filters.groups ?? []) {
    params.append("group", group);
  }

  for (const tierId of filters.membershipTierIds ?? []) {
    params.append("membership_tier_id", tierId);
  }

  if (filters.isStudent !== undefined) {
    params.set("is_student", String(filters.isStudent));
  }

  if (pagination) {
    params.set("limit", String(pagination.limit));
    params.set("offset", String(pagination.offset));
  }

  return params;
}

export async function fetchUsers(
  appliedSearch: AppliedSearch,
  filters: AdminUserFilters,
  pagination: AdminPagination,
  signal?: AbortSignal,
): Promise<UsersResponse> {
  const response = await apiClient.get<UsersResponse>("/admin/users", {
    params: buildAdminUserParams(appliedSearch, filters, pagination),
    signal,
  });

  return response.data;
}

export async function fetchAdminMembershipTierOptions(
  signal?: AbortSignal,
): Promise<AdminMembershipTierOption[]> {
  const response = await apiClient.get<AdminMembershipTierOption[]>("/admin/membership-tiers", {
    signal,
  });

  return response.data ?? [];
}

export async function fetchUser(userId: string, signal?: AbortSignal): Promise<User> {
  const response = await apiClient.get<UserResponse>(`/admin/users/${userId}`, { signal });

  return response.data.user;
}

export async function fetchUserMemberships(
  userId: string,
  signal?: AbortSignal,
): Promise<Membership[]> {
  const response = await apiClient.get<Membership[]>(`/admin/users/${userId}/memberships`, {
    signal,
  });

  return response.data ?? [];
}

export async function fetchEligibleMembershipsForUser(
  userId: string,
  signal?: AbortSignal,
): Promise<EligibleMembershipTier[]> {
  const response = await apiClient.get<EligibleMembershipTier[]>(
    `/admin/memberships/eligible/${userId}`,
    { signal },
  );

  return response.data ?? [];
}

export async function addOfflineMembership(
  userId: string,
  body: AddOfflineMembershipRequest,
): Promise<void> {
  await apiClient.post(`/admin/membership/add/${userId}`, body);
}

export async function fetchMembershipInvitations(
  signal?: AbortSignal,
): Promise<MembershipInvitation[]> {
  const response = await apiClient.get<MembershipInvitation[]>("/admin/membership-invitations", {
    signal,
  });

  return response.data ?? [];
}

export async function createMembershipInvitation(
  body: CreateMembershipInvitationRequest,
): Promise<MembershipInvitation> {
  const response = await apiClient.post<MembershipInvitation>(
    "/admin/membership-invitations",
    body,
  );

  return response.data;
}

export async function deleteMembershipInvitation(id: string): Promise<void> {
  await apiClient.delete(`/admin/membership-invitations/${id}`);
}

export async function updateUser(userId: string, body: UpdateUserRequest): Promise<User> {
  const response = await apiClient.patch<UserResponse>(`/admin/users/${userId}`, body);

  return response.data.user;
}

export async function exportUsersCSV(
  appliedSearch: AppliedSearch,
  filters: AdminUserFilters,
  signal?: AbortSignal,
): Promise<Blob> {
  const response = await apiClient.get<Blob>("/admin/users/export", {
    params: buildAdminUserParams(appliedSearch, filters),
    responseType: "blob",
    signal,
  });

  return response.data;
}

export function buildAdminAuditLogParams(
  actorName?: string,
  pagination?: AdminPagination,
): Record<string, string | number> {
  const params: Record<string, string | number> = {};

  if (actorName?.trim()) {
    params.actor_name = actorName.trim();
  }

  if (pagination) {
    params.limit = pagination.limit;
    params.offset = pagination.offset;
  }

  return params;
}

export async function fetchAuditLogs(
  actorName?: string,
  pagination?: AdminPagination,
  signal?: AbortSignal,
): Promise<AuditLogResponse> {
  const response = await apiClient.get<AuditLogResponse>("/admin/audit-logs", {
    params: buildAdminAuditLogParams(actorName, pagination),
    signal,
  });
  return response.data;
}

export async function exportAuditLogsCSV(actorName?: string, signal?: AbortSignal): Promise<Blob> {
  const response = await apiClient.get<Blob>("/admin/audit-logs/export", {
    params: buildAdminAuditLogParams(actorName),
    responseType: "blob",
    signal,
  });

  return response.data;
}

export function downloadCSVBlob(blob: Blob, filename = "users.csv") {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}
