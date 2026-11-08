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

// LoungeSessionHistoryDTO is a past or active lounge session shown in a user's history.
// DurationSeconds is nil while the session is still active (not checked out).
type LoungeSessionHistoryDTO struct {
	ID              string     `json:"id"`
	Game            *string    `json:"game"`
	StationLabel    string     `json:"station_label"`
	CheckInTime     time.Time  `json:"check_in_time"`
	CheckOutTime    *time.Time `json:"check_out_time"`
	DurationSeconds *int64     `json:"duration_seconds"`
}
