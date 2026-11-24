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
	// ActiveSessionCheckInTime is when the active session started; nil when the station is free.
	ActiveSessionCheckInTime *time.Time `json:"active_session_check_in_time"`
	InUse                    bool       `json:"in_use"`
}

type UpdateLoungeStationRequestDTO struct {
	Label  *string `json:"label"`
	Closed *bool   `json:"closed"`
}

type LoungeSessionDTO struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	MembershipID *string    `json:"membership_id"`
	StationID    string     `json:"station_id"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time"`
	CheckedInBy  string     `json:"checked_in_by"`
	CheckedOutBy *string    `json:"checked_out_by"`
	Game         *string    `json:"game"`
}

type CheckoutSessionRequestDTO struct {
	CheckedOutBy *string `json:"checked_out_by"`
}

// LoungeSessionDetailsDTO describes a lounge session with the member and staff
// details needed by the lounge map's station panel.
type LoungeSessionDetailsDTO struct {
	ID                  string     `json:"id"`
	StationID           string     `json:"station_id"`
	MembershipID        *string    `json:"membership_id"`
	MembershipTierTitle *string    `json:"membership_tier_title"`
	CheckInTime         time.Time  `json:"check_in_time"`
	CheckOutTime        *time.Time `json:"check_out_time"`
	Game                *string    `json:"game"`
	UserID              string     `json:"user_id"`
	UserFullName        string     `json:"user_full_name"`
	UserEmail           string     `json:"user_email"`
	UserStudentID       *string    `json:"user_student_id"`
	UserAvatarURL       *string    `json:"user_avatar_url"`
	CheckedInBy         string     `json:"checked_in_by"`
	CheckedInByName     string     `json:"checked_in_by_name"`
}
