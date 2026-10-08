"use client";

import { ChevronRight, Loader2, RefreshCw, Save, Trash2 } from "lucide-react";
import {
  FaInstagram,
  FaLinkedin,
  FaTiktok,
  FaTwitch,
  FaXTwitter,
  FaYoutube,
} from "react-icons/fa6";
import type { IconType } from "react-icons";
import { useState, type FormEvent } from "react";
import { toast } from "sonner";
import { ActionButton } from "@/components/action-button";
import { SurfacePanel } from "@/components/surface-panel";
import {
  useDeleteExecSocialLink,
  useExecProfile,
  useUpdateExecSocialLink,
  useUpdateExecTitle,
} from "@/lib/exec-profile.hook";
import type { ExecSocialPlatform } from "@/lib/types/exec-profile.types";

const FIELD_CLASS_NAME =
  "h-11 w-full border border-brand-border bg-brand-surface px-3 text-sm text-brand-text outline-none transition placeholder:text-brand-text-subtle/70 focus:border-brand-primary focus:ring-2 focus:ring-brand-primary/30 disabled:cursor-not-allowed disabled:opacity-60";

const SOCIAL_PLATFORMS: {
  platform: ExecSocialPlatform;
  label: string;
  placeholder: string;
  icon: IconType;
  iconClassName: string;
}[] = [
  {
    platform: "instagram",
    label: "Instagram",
    placeholder: "https://instagram.com/username",
    icon: FaInstagram,
    iconClassName: "text-pink-300",
  },
  {
    platform: "x",
    label: "X",
    placeholder: "https://x.com/username",
    icon: FaXTwitter,
    iconClassName: "text-brand-text",
  },
  {
    platform: "twitch",
    label: "Twitch",
    placeholder: "https://twitch.tv/username",
    icon: FaTwitch,
    iconClassName: "text-violet-300",
  },
  {
    platform: "youtube",
    label: "YouTube",
    placeholder: "https://youtube.com/@channel",
    icon: FaYoutube,
    iconClassName: "text-red-400",
  },
  {
    platform: "tiktok",
    label: "TikTok",
    placeholder: "https://tiktok.com/@username",
    icon: FaTiktok,
    iconClassName: "text-cyan-300",
  },
  {
    platform: "linkedin",
    label: "LinkedIn",
    placeholder: "https://linkedin.com/in/username",
    icon: FaLinkedin,
    iconClassName: "text-sky-400",
  },
];

function isValidWebUrl(value: string) {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

type SocialLinkRowProps = {
  platform: ExecSocialPlatform;
  label: string;
  placeholder: string;
  initialValue: string;
  icon: IconType;
  iconClassName: string;
};

function SocialLinkRow({
  platform,
  label,
  placeholder,
  initialValue,
  icon: Icon,
  iconClassName,
}: SocialLinkRowProps) {
  const [url, setUrl] = useState(initialValue);
  const { mutateAsync: updateLink, isPending: isSaving } = useUpdateExecSocialLink();
  const { mutateAsync: deleteLink, isPending: isDeleting } = useDeleteExecSocialLink();
  const isBusy = isSaving || isDeleting;
  const trimmedUrl = url.trim();
  const hasSavedLink = initialValue.length > 0;
  const isDirty = trimmedUrl !== initialValue;

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!trimmedUrl) {
      toast.error(`Enter a ${label} URL`);
      return;
    }
    if (!isValidWebUrl(trimmedUrl)) {
      toast.error(`Enter a complete ${label} URL`, {
        description: "Include https:// at the beginning.",
      });
      return;
    }

    try {
      await updateLink({ platform, url: trimmedUrl });
      toast.success(`${label} link saved`);
    } catch {
      // The shared API client displays the server error.
    }
  };

  const handleDelete = async () => {
    try {
      await deleteLink(platform);
      setUrl("");
      toast.success(`${label} link removed`);
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="grid gap-2 sm:grid-cols-[8rem_minmax(0,1fr)_auto] sm:items-center"
    >
      <label
        htmlFor={`exec-social-${platform}`}
        className="flex items-center gap-2 text-sm font-medium text-brand-text-muted"
      >
        <span
          aria-hidden="true"
          className="flex size-7 shrink-0 items-center justify-center border border-brand-border bg-white/5"
        >
          <Icon className={`size-4 ${iconClassName}`} />
        </span>
        {label}
      </label>
      <input
        id={`exec-social-${platform}`}
        type="url"
        inputMode="url"
        autoComplete="url"
        value={url}
        onChange={(event) => setUrl(event.target.value)}
        placeholder={placeholder}
        disabled={isBusy}
        className={FIELD_CLASS_NAME}
      />
      <div className="flex gap-2 sm:justify-end">
        {hasSavedLink ? (
          <ActionButton
            onClick={handleDelete}
            disabled={isBusy}
            loading={isDeleting}
            icon={<Trash2 aria-hidden="true" className="size-4" />}
            loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
            aria-label={`Remove ${label} link`}
            className="border-red-400/40 px-3 text-red-100 hover:border-red-300 hover:bg-red-400/10"
          >
            Remove
          </ActionButton>
        ) : null}
        <ActionButton
          type="submit"
          disabled={!trimmedUrl || !isDirty || isBusy}
          loading={isSaving}
          icon={<Save aria-hidden="true" className="size-4" />}
          loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
          className="min-w-24 border-brand-primary bg-brand-primary hover:border-brand-primary-hover hover:bg-brand-primary-hover"
        >
          Save
        </ActionButton>
      </div>
    </form>
  );
}

function ExecTitleForm({ initialTitle }: { initialTitle: string }) {
  const [title, setTitle] = useState(initialTitle);
  const { mutateAsync: updateTitle, isPending: isSavingTitle } = useUpdateExecTitle();

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedTitle = title.trim();

    if (!trimmedTitle) {
      toast.error("Enter a role title");
      return;
    }

    try {
      await updateTitle(trimmedTitle);
      setTitle(trimmedTitle);
      toast.success("Executive title saved");
    } catch {
      // The shared API client displays the server error.
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-start"
    >
      <label className="flex min-w-0 flex-col gap-1.5 text-sm text-brand-text-subtle">
        <span className="font-medium text-brand-text-muted">Role title</span>
        <input
          type="text"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          maxLength={100}
          disabled={isSavingTitle}
          className={FIELD_CLASS_NAME}
          placeholder="e.g. VP Events"
        />
        <span className="text-xs">Use the title members should see beneath your name.</span>
      </label>
      <ActionButton
        type="submit"
        disabled={!title.trim() || title.trim() === initialTitle}
        loading={isSavingTitle}
        icon={<Save aria-hidden="true" className="size-4" />}
        loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
        className="h-11 border-brand-primary bg-brand-primary hover:border-brand-primary-hover hover:bg-brand-primary-hover lg:mt-[1.625rem]"
      >
        Save title
      </ActionButton>
    </form>
  );
}

export function ExecProfilePanel() {
  const { data: execProfile, isPending, isError, isFetching, refetch } = useExecProfile(true);

  return (
    <section id="exec-profile" className="mt-6 scroll-mt-28" aria-labelledby="exec-profile-heading">
      <SurfacePanel className="overflow-hidden bg-transparent">
        <details className="group/exec-profile">
          <summary className="flex cursor-pointer list-none items-center gap-3 bg-brand-primary/10 px-5 py-4 transition hover:bg-brand-primary/15 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-brand-primary group-open/exec-profile:border-b group-open/exec-profile:border-brand-border [&::-webkit-details-marker]:hidden">
            <ChevronRight
              aria-hidden="true"
              className="size-4 shrink-0 text-brand-text-subtle transition-transform group-open/exec-profile:rotate-90"
            />
            <span>
              <span
                id="exec-profile-heading"
                className="block text-base font-semibold text-brand-text"
              >
                Executive profile
              </span>
              <span className="mt-1 block text-sm text-brand-text-subtle">
                Control how your role and social links appear on the public executive roster.
              </span>
            </span>
          </summary>

          {isPending ? (
            <div className="flex items-center gap-3 px-5 py-8 text-sm text-brand-text-muted">
              <Loader2 aria-hidden="true" className="size-4 animate-spin" />
              Loading executive profile
            </div>
          ) : isError || !execProfile ? (
            <div
              role="alert"
              className="flex flex-col gap-3 px-5 py-6 sm:flex-row sm:items-center sm:justify-between"
            >
              <p className="text-sm text-brand-text-muted">
                Your executive profile could not be loaded.
              </p>
              <ActionButton
                onClick={() => void refetch()}
                loading={isFetching}
                icon={<RefreshCw aria-hidden="true" className="size-4" />}
                loadingIcon={<Loader2 aria-hidden="true" className="size-4 animate-spin" />}
              >
                Try again
              </ActionButton>
            </div>
          ) : (
            <div className="grid gap-7 p-5 sm:p-6">
              <ExecTitleForm key={execProfile.title} initialTitle={execProfile.title} />

              <div className="border-t border-brand-border pt-6">
                <div className="mb-4">
                  <h4 className="text-sm font-semibold text-brand-text">Social links</h4>
                  <p className="mt-1 text-sm text-brand-text-subtle">
                    Add only the profiles you want shown publicly.
                  </p>
                </div>
                <div className="grid gap-3">
                  {SOCIAL_PLATFORMS.map((social) => (
                    <SocialLinkRow
                      key={`${social.platform}:${execProfile.social_links.find(({ platform }) => platform === social.platform)?.url ?? ""}`}
                      {...social}
                      initialValue={
                        execProfile.social_links.find(
                          ({ platform }) => platform === social.platform,
                        )?.url ?? ""
                      }
                    />
                  ))}
                </div>
              </div>
            </div>
          )}
        </details>
      </SurfacePanel>
    </section>
  );
}
