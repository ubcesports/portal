package dto

type LoungeStationDTO struct {
	ID              string            `json:"id"`
	Code            string            `json:"code"`
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
