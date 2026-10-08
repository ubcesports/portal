import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "./client";

import type {
  ExecProfileResponse,
  ExecSocialLink,
  ExecSocialPlatform,
} from "@/lib/types/exec-profile.types";

const EXEC_PROFILE_QUERY_KEY = ["exec-profile", "me"] as const;

export function useExecProfile(enabled: boolean) {
  return useQuery({
    queryKey: EXEC_PROFILE_QUERY_KEY,
    queryFn: async ({ signal }) => {
      const { data } = await apiClient.get<ExecProfileResponse>("/exec-profile", { signal });
      return data.exec_profile;
    },
    enabled,
  });
}

export function useUpdateExecTitle() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (title: string) => {
      await apiClient.patch("/exec-profile/title", { title });
      return title;
    },
    onSuccess: (title) => {
      queryClient.setQueryData<ExecProfileResponse["exec_profile"]>(
        EXEC_PROFILE_QUERY_KEY,
        (profile) => (profile ? { ...profile, title } : profile),
      );
    },
  });
}

export function useUpdateExecSocialLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (link: ExecSocialLink) => {
      await apiClient.put("/exec-profile/social-links", link);
      return link;
    },
    onSuccess: (link) => {
      queryClient.setQueryData<ExecProfileResponse["exec_profile"]>(
        EXEC_PROFILE_QUERY_KEY,
        (profile) =>
          profile
            ? {
                ...profile,
                social_links: [
                  ...profile.social_links.filter(({ platform }) => platform !== link.platform),
                  link,
                ],
              }
            : profile,
      );
    },
  });
}

export function useDeleteExecSocialLink() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (platform: ExecSocialPlatform) => {
      await apiClient.delete("/exec-profile/social-links", { data: { platform } });
      return platform;
    },
    onSuccess: (deletedPlatform) => {
      queryClient.setQueryData<ExecProfileResponse["exec_profile"]>(
        EXEC_PROFILE_QUERY_KEY,
        (profile) =>
          profile
            ? {
                ...profile,
                social_links: profile.social_links.filter(
                  ({ platform }) => platform !== deletedPlatform,
                ),
              }
            : profile,
      );
    },
  });
}
