package dto

import "time"

type LoungeStationDTO struct {
	ID              string            `json:"id"`
	Label           string            `json:"label"`
	StationType     LoungeStationType `json:"station_type"`
	Closed          bool              `json:"closed"`
	Row             int16             `json:"row"`
	Col             int16             `json:"col"`
	Width           int16             `json:"width"`
	Height          int16             `json:"height"`
	ActiveSessionID *string           `json:"active_session_id"`
	InUse           bool              `json:"in_use"`
}

type UpdateLoungeStationRequestDTO struct {
	Label  *string `json:"label"`
	Closed *bool   `json:"closed"`
}

type LoungeMemberLookupDTO struct {
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
	Email     string  `json:"email"`
	StudentID *string `json:"student_id"`
}

type LoungeSessionDTO struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	MembershipID string    `json:"membership_id"`
	StationID    string    `json:"station_id"`
	CheckInTime  time.Time `json:"check_in_time"`
	CheckedInBy  string    `json:"checked_in_by"`
	Game         *string   `json:"game,omitempty"`
}

type CreateLoungeSessionRequest struct {
	StudentID string  `json:"student_id"`
	StationID string  `json:"station_id"`
	Game      *string `json:"game"`
}
