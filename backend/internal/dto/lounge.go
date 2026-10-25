package dto

import "time"

// Lounge sessions

type LoungeSessionDTO struct {
	ID           string     `json:"id"`
	Game         *string    `json:"game"`
	StationLabel string     `json:"station_label"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time"`
}
