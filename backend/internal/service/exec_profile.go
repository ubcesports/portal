package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

type ExecProfileService struct {
	execProfileRepository *repository.ExecProfileRepository
}

func NewExecProfileService(execProfileRepository *repository.ExecProfileRepository) *ExecProfileService {
	return &ExecProfileService{execProfileRepository: execProfileRepository}
}

func (s *ExecProfileService) GetExecProfiles(ctx context.Context) ([]*dto.ExecProfileDTO, error) {
	rows, err := s.execProfileRepository.GetExecProfiles(ctx)
	if err != nil {
		return nil, err
	}

	var execProfiles []*dto.ExecProfileDTO
	for _, row := range rows {
		execProfile, err := buildExecProfile(row)
		if err != nil {
			return nil, err
		}
		execProfiles = append(execProfiles, execProfile)
	}

	return execProfiles, nil
}

func (s *ExecProfileService) UpdateExecSocialLink(ctx context.Context, userId string, platform db.ExecSocialPlatformType, url string) error {
	if err := s.execProfileRepository.UpdateExecSocialLink(ctx, userId, platform, url); err != nil {
		return fmt.Errorf("update exec social link: %w", err)
	}
	return nil
}

func (s *ExecProfileService) DeleteExecSocialLink(ctx context.Context, userId string, platform db.ExecSocialPlatformType) error {
	if err := s.execProfileRepository.DeleteExecSocialLink(ctx, userId, platform); err != nil {
		return fmt.Errorf("delete exec social link: %w", err)
	}
	return nil
}

func (s *ExecProfileService) UpdateExecProfileTitle(ctx context.Context, userId string, title string) (db.GetExecProfileByUserIDRow, error) {
	updated, err := s.execProfileRepository.UpdateExecProfileByUserID(ctx, userId, title, 0, "")
	if err != nil {
		return db.GetExecProfileByUserIDRow{}, fmt.Errorf("update exec profile title: %w", err)
	}
	return updated, nil
}

/*
	Private functions
*/

func buildExecProfile(row db.GetExecProfilesRow) (*dto.ExecProfileDTO, error) {
	// Unmarshal the social links JSON into a slice of ExecProfileSocialLinkDTO
	socialLinksJSON, err := json.Marshal(row.SocialLinks)
	if err != nil {
		return nil, fmt.Errorf("marshal social links: %w", err)
	}
	var socialLinks []dto.ExecProfileSocialLinkDTO
	if err := json.Unmarshal(socialLinksJSON, &socialLinks); err != nil {
		return nil, fmt.Errorf("unmarshal social links: %w", err)
	}

	// Build exec profile DTO
	execProfile := &dto.ExecProfileDTO{
		FullName:     row.FullName,
		AvatarURL:    util.TextPointer(row.AvatarUrl),
		Title:        row.Title,
		DisplayGroup: dto.ExecDisplayGroupType(row.DisplayGroup),
		SocialLinks:  socialLinks,
	}

	return execProfile, nil
}
