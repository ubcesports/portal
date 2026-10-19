package dto

type ExecProfileDTO struct {
	FullName     string                     `json:"full_name"`
	AvatarURL    *string                    `json:"avatar_url"`
	Title        string                     `json:"title"`
	DisplayGroup ExecDisplayGroupType       `json:"display_group"`
	SocialLinks  []ExecProfileSocialLinkDTO `json:"social_links"`
}

type ExecProfileSocialLinkDTO struct {
	Platform ExecSocialPlatformType `json:"platform"`
	URL      string                 `json:"url"`
}
