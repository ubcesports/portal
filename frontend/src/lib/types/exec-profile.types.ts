export type ExecSocialPlatform = "instagram" | "x" | "twitch" | "youtube" | "tiktok" | "linkedin";

export type ExecDisplayGroup =
  "president" | "board" | "central_director" | "game_director" | "executive";

export type ExecSocialLink = {
  platform: ExecSocialPlatform;
  url: string;
};

export type ExecProfile = {
  full_name: string;
  avatar_url: string | null;
  title: string;
  display_group: ExecDisplayGroup;
  social_links: ExecSocialLink[];
};

export type ExecProfileResponse = {
  exec_profile: ExecProfile;
};
