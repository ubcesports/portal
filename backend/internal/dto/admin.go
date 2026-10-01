package dto

import "time"

type AdminUserDTO struct {
	ProfileDTO
	ActiveMemberships []AdminActiveMembershipSummary `json:"active_memberships"`
}

type AdminActiveMembershipSummary struct {
	TierTitle string `json:"tier_title"`
}

type AdminMembershipTierOption struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ProgramName string `json:"program_name"`
}

type AdminAuditLogResponse struct {
	Actor       AdminAuditLogActor       `json:"actor"`
	OccuredAt   time.Time                `json:"occured_at"`
	Action      string                   `json:"action"`
	Description *string                  `json:"description"`
	Outcome     AdminAuditLogOutcomeType `json:"outcome"`
	RequestId   string                   `json:"request_id"`
	TargetUser  *AdminAuditLogActor      `json:"target_user"`
}

// AdminUpdateUserRequest is the body of PATCH /admin/users/{id}. Every field is
// optional and only the ones present are applied.
type AdminUpdateUserRequest struct {
	FullName           *string     `json:"full_name"`
	StudentID          *string     `json:"student_id"`
	IsStudent          *bool       `json:"is_student"`
	GroupsAdd          []GroupType `json:"groups_add"`
	GroupsRemove       []GroupType `json:"groups_remove"`
	Role               *RoleType   `json:"role"`
	CancelMembershipId *string     `json:"cancel_membership_id"`
}

type AdminUpdateExecProfileRequest struct {
	Title        *string    `json:"title"`
	DisplayOrder *int32     `json:"display_order"`
	DisplayGroup *GroupType `json:"display_group"`
}

type AdminAuditLogActor struct {
	ActorUserId    string `json:"actor_user_id"`
	ActorFullName  string `json:"actor_full_name"`
	ActorAvatarURL string `json:"actor_avatar_url"`
}

type AdminAddMembershipToUserRequest struct {
	TierId        string            `json:"tier_id"`
	PaymentMethod PaymentMethodType `json:"payment_method"`
}
