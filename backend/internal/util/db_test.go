package util

import (
	"testing"

	"github.com/ubcesports/memberships/internal/database/db"
	"github.com/ubcesports/memberships/internal/dto"
)

func TestToNullExecDisplayGroupType(t *testing.T) {
	if got := ToNullExecDisplayGroupType(nil); got.Valid {
		t.Fatalf("expected nil display group to remain null, got %#v", got)
	}

	value := dto.ExecDisplayGroupTypeCentralDirector
	got := ToNullExecDisplayGroupType(&value)
	if !got.Valid || got.ExecDisplayGroupType != db.ExecDisplayGroupTypeCentralDirector {
		t.Fatalf("expected central director display group, got %#v", got)
	}
}
