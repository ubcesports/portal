package dto

import "time"

// Request

type CreateRoleRequest struct {
	Name           string   `json:"name"`
	Description    *string  `json:"description"`
	Color          *string  `json:"color"`
	DisplayOrder   *int32   `json:"display_order"`
	PermissionKeys []string `json:"permission_keys"`
}

type UpdateRoleRequest struct {
	Name           *string   `json:"name"`
	Description    *string   `json:"description"`
	Color          *string   `json:"color"`
	DisplayOrder   *int32    `json:"display_order"`
	PermissionKeys *[]string `json:"permission_keys"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id"`
}

// Response

type RoleResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	Color        *string   `json:"color"`
	DisplayOrder int32     `json:"display_order"`
	CreatedBy    *RoleUser `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoleDetailsResponse struct {
	RoleResponse
	PermissionKeys []string   `json:"permission_keys"`
	UsersWithRole  []RoleUser `json:"users_with_role"`
}

type RoleUser struct {
	UserID         string  `json:"user_id"`
	Name           string  `json:"name"`
	ProfilePicture *string `json:"profile_picture"`
}

type UserRolePermissions struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Color          *string  `json:"color"`
	PermissionKeys []string `json:"permission_keys"`
}
