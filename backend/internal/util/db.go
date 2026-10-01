package util

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
)

func GetValidatedUUID(uuid string) (pgtype.UUID, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(uuid); err != nil {
		return pgtype.UUID{}, err
	}
	return pgUUID, nil
}

func TextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func TimestampPointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func ToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func ToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func ToNullGroupType(value *dto.GroupType) db.NullGroupType {
	if value == nil {
		return db.NullGroupType{}
	}
	return db.NullGroupType{
		GroupType: db.GroupType(*value),
		Valid:     true,
	}
}
