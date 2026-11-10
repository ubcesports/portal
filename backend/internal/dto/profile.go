package dto

import "time"

type ProfileDTO struct {
	ID                    string      `json:"id"`
	Email                 string      `json:"email"`
	StudentID             *string     `json:"student_id"`
	Role                  RoleType    `json:"role"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
	FullName              string      `json:"full_name"`
	EmailVerifiedAt       *time.Time  `json:"email_verified_at"`
	IsStudent             bool        `json:"is_student"`
	OnboardingCompletedAt *time.Time  `json:"onboarding_completed_at"`
	AvatarURL             *string     `json:"avatar_url"`
	Groups                []GroupType `json:"groups"`
	// Roles are the custom roles assigned to the user, with the permissions each
	// grants. Only GET /profile fills this in; it is null elsewhere.
	Roles []UserRolePermissions `json:"roles"`
}

type OnboardUserRequest struct {
	FullName   string  `json:"full_name"`
	IsStudent  bool    `json:"is_student"`
	StudentID  *string `json:"student_id"`
	InviteCode *string `json:"invite_code"`
}
