package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/mailer"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

// Errors
var (
	ErrValidation = errors.New("Validation error")
	ErrConflict   = errors.New("Conflict")
	ErrNotFound   = errors.New("Not found")
)

var studentIDRegex = regexp.MustCompile(`^\d{8}$`) // Regex to ensure student id is an 8 digit number, which all ubc student ids are

type ProfileService struct {
	profileRepository           *repository.ProfileRepository
	membershipInvitationService *MembershipInvitationService
}

func NewProfileService(
	profileRepository *repository.ProfileRepository,
	membershipInvitationService *MembershipInvitationService,
) *ProfileService {
	return &ProfileService{
		profileRepository:           profileRepository,
		membershipInvitationService: membershipInvitationService,
	}
}

func (s *ProfileService) GetProfileByUserID(ctx context.Context, userID string) (*dto.ProfileDTO, error) {
	if err := s.profileRepository.EnsureMemberGroupForUser(ctx, userID); err != nil {
		return nil, err
	}

	row, err := s.profileRepository.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := buildProfile(row)
	if profile.OnboardingCompletedAt != nil && s.membershipInvitationService != nil {
		if _, err := s.membershipInvitationService.RedeemForUser(ctx, profile); err != nil {
			slog.ErrorContext(ctx, "redeem membership invitation failed",
				"error", err,
				"user_id", profile.ID,
			)
		}
	}
	return profile, nil
}

func (s *ProfileService) OnboardUser(ctx context.Context, userId string, onboardUserRequest dto.OnboardUserRequest) error {
	fullName := strings.TrimSpace(onboardUserRequest.FullName)
	if fullName == "" {
		return fmt.Errorf("%w: full name is required", ErrValidation)
	}

	var studentID string

	// Get user profile
	user, err := s.GetProfileByUserID(ctx, userId)
	if err != nil {
		return err
	}

	// Only allow non-onboarded users to onboard
	if user.OnboardingCompletedAt != nil {
		return fmt.Errorf("%w: current user is already onboarded!", ErrConflict)
	}

	if onboardUserRequest.IsStudent {

		// Students should provide their id
		if onboardUserRequest.StudentID == nil {
			return fmt.Errorf("%w: student ID is required for students", ErrValidation)
		}

		studentID = strings.TrimSpace(*onboardUserRequest.StudentID)
		if studentID == "" {
			return fmt.Errorf("%w: student ID is required for students", ErrValidation)
		}

		// ID should be 8 digit number
		if !studentIDRegex.MatchString(studentID) {
			return fmt.Errorf("%w: student ID must be an 8 digit number", ErrValidation)
		}
	} else {
		studentID = util.GenerateNonStudentID()
	}

	// Allow execs to get exec group through a code
	expectedCode := os.Getenv("EXEC_ONBOARDING_CODE")
	isExec := onboardUserRequest.InviteCode != nil &&
		expectedCode != "" &&
		strings.TrimSpace(*onboardUserRequest.InviteCode) == expectedCode

	if err := s.profileRepository.OnboardUserByUserId(
		ctx,
		userId,
		fullName,
		onboardUserRequest.IsStudent,
		studentID,
		isExec,
	); err != nil {
		return err
	}

	updatedRow, err := s.profileRepository.GetProfileByUserID(ctx, userId)
	if err != nil {
		return err
	}
	updatedUser := buildProfile(updatedRow)
	if s.membershipInvitationService != nil {
		if _, err := s.membershipInvitationService.RedeemForUser(ctx, updatedUser); err != nil {
			slog.ErrorContext(ctx, "redeem membership invitation after onboarding failed",
				"error", err,
				"user_id", updatedUser.ID,
			)
		}
	}

	s.sendWelcomeEmail(ctx, updatedUser)
	return nil
}

// sendWelcomeEmail fires the one-time signup welcome email. Onboarding has
// already succeeded by this point, so a failure here is logged and swallowed
// rather than surfaced to the user.
func (s *ProfileService) sendWelcomeEmail(ctx context.Context, user *dto.ProfileDTO) {
	data := welcomeEmail()

	html, err := mailer.RenderEmail(data)
	if err != nil {
		slog.Error("render welcome email failed", "error", err, "user_id", user.ID)
		return
	}

	mailer.SendEmailAsync(
		[]string{user.Email},
		data.Heading,
		html,
		middleware.GetReqID(ctx),
		user.ID,
	)
}

// welcomeEmail builds the one-time signup welcome email's content.
func welcomeEmail() mailer.EmailData {
	heading := "Welcome to UBCEA!"
	return mailer.EmailData{
		Title:      heading,
		Heading:    heading,
		Subheading: "Thanks for signing up. You're now part of the UBC Esports Association community — the next step is picking a membership so you can access events, the lounge, and everything else we run throughout the year.",
		CTAText:    "View membership pricing",
		CTAURL:     mailer.FrontendURL() + "/pricing",
	}
}

/*
	Private functions
*/

func buildProfile(row db.GetProfileByUserIDRow) *dto.ProfileDTO {
	profile := &dto.ProfileDTO{
		ID:                    row.ID.String(),
		Email:                 row.Email,
		StudentID:             util.TextPointer(row.StudentID),
		Role:                  dto.RoleType(row.Role),
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
		FullName:              row.FullName,
		EmailVerifiedAt:       util.TimestampPointer(row.EmailVerifiedAt),
		IsStudent:             row.IsStudent,
		OnboardingCompletedAt: util.TimestampPointer(row.OnboardingCompletedAt),
		AvatarURL:             util.TextPointer(row.AvatarUrl),
		Groups:                make([]dto.GroupType, 0, len(row.Groups)),
	}

	for _, group := range row.Groups {
		profile.Groups = append(profile.Groups, dto.GroupType(group))
	}

	return profile
}
