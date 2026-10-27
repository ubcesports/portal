package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
	"github.com/ubcesports/memberships/internal/util"
)

var (
	ErrInvalidLoungeSession  = errors.New("Invalid lounge session")
	ErrNoLoungeAccess        = errors.New("User does not have lounge access")
	ErrActiveSessionExists   = errors.New("User already has an active lounge session")
	ErrStationClosed         = errors.New("Lounge station is closed")
	ErrStationInUse          = errors.New("Lounge station is already in use")
	ErrLoungeStationNotFound = errors.New("Lounge station not found")
)

type LoungeService struct {
	loungeRepo *repository.LoungeRepository
}

func NewLoungeService(loungeRepo *repository.LoungeRepository) *LoungeService {
	return &LoungeService{loungeRepo: loungeRepo}
}

func (s *LoungeService) CreateLoungeSession(
	ctx context.Context,
	studentID string,
	stationID string,
	game *string,
	checkedInBy string,
) (*dto.LoungeSessionDTO, error) {
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return nil, ErrInvalidLoungeSession
	}
	if _, err := util.GetValidatedUUID(stationID); err != nil {
		return nil, ErrInvalidLoungeSession
	}
	if _, err := util.GetValidatedUUID(checkedInBy); err != nil {
		return nil, ErrInvalidLoungeSession
	}
	if game != nil {
		trimmedGame := strings.TrimSpace(*game)
		if trimmedGame == "" {
			game = nil
		} else {
			game = &trimmedGame
		}
	}

	membership, err := s.loungeRepo.GetEligibleMembershipByStudentID(ctx, studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoLoungeAccess
	}
	if err != nil {
		return nil, fmt.Errorf("get lounge membership for check-in: %w", err)
	}

	userID := membership.UserID.String()
	hasActiveSession, err := s.loungeRepo.HasActiveSessionForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check member's active lounge session: %w", err)
	}
	if hasActiveSession {
		return nil, ErrActiveSessionExists
	}

	station, err := s.loungeRepo.GetLoungeStationByIDWithActiveSession(ctx, stationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrLoungeStationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get lounge station for check-in: %w", err)
	}
	if station.Closed {
		return nil, ErrStationClosed
	}
	if station.ActiveSessionID.Valid {
		return nil, ErrStationInUse
	}

	session, err := s.loungeRepo.CreateLoungeSession(ctx, repository.CreateLoungeSessionParams{
		StudentID:    studentID,
		MembershipID: membership.MembershipID.String(),
		StationID:    stationID,
		CheckedInBy:  checkedInBy,
		Game:         game,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "lounge_sessions_one_active_per_user":
				return nil, ErrActiveSessionExists
			case "lounge_sessions_one_active_per_station":
				return nil, ErrStationInUse
			}
		}
		return nil, fmt.Errorf("create lounge session: %w", err)
	}

	return &dto.LoungeSessionDTO{
		ID:           session.ID.String(),
		UserID:       session.UserID.String(),
		MembershipID: session.MembershipID.String(),
		StationID:    session.StationID.String(),
		CheckInTime:  session.CheckInTime.Time,
		CheckedInBy:  session.CheckedInBy.String(),
		Game:         util.TextPointer(session.Game),
	}, nil
}

func (s *LoungeService) GetLoungeStations(ctx context.Context) ([]dto.LoungeStationDTO, error) {
	rows, err := s.loungeRepo.GetLoungeStations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.LoungeStationDTO, 0, len(rows))
	for _, row := range rows {
		var activeSessionID *string
		inUse := false
		if row.ActiveSessionID.Valid {
			sessionIDStr := row.ActiveSessionID.String()
			activeSessionID = &sessionIDStr
			inUse = true
		}

		result = append(result, dto.LoungeStationDTO{
			ID:              row.ID.String(),
			Label:           row.Label,
			StationType:     dto.LoungeStationType(row.StationType),
			Closed:          row.Closed,
			Row:             row.GridRow,
			Col:             row.GridColumn,
			Width:           row.Width,
			Height:          row.Height,
			ActiveSessionID: activeSessionID,
			InUse:           inUse,
		})
	}

	return result, nil
}

func (s *LoungeService) UpdateLoungeStation(ctx context.Context, id string, req dto.UpdateLoungeStationRequestDTO) (*dto.LoungeStationDTO, error) {
	current, err := s.loungeRepo.GetLoungeStationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if label or closed are changing
	var updateLabel *string
	if req.Label != nil && *req.Label != current.Label {
		updateLabel = req.Label
	}

	// If closed is already the requested status, do nothing for closed
	var updateClosed *bool
	if req.Closed != nil && *req.Closed != current.Closed {
		updateClosed = req.Closed
	}

	// If neither field changed, do nothing and return the current state
	if updateLabel == nil && updateClosed == nil {
		currentWithSession, err := s.loungeRepo.GetLoungeStationByIDWithActiveSession(ctx, id)
		if err != nil {
			return nil, err
		}
		var activeSessionID *string
		inUse := false
		if currentWithSession.ActiveSessionID.Valid {
			s := currentWithSession.ActiveSessionID.String()
			activeSessionID = &s
			inUse = true
		}
		return &dto.LoungeStationDTO{
			ID:              currentWithSession.ID.String(),
			Label:           currentWithSession.Label,
			StationType:     dto.LoungeStationType(currentWithSession.StationType),
			Closed:          currentWithSession.Closed,
			Row:             currentWithSession.GridRow,
			Col:             currentWithSession.GridColumn,
			Width:           currentWithSession.Width,
			Height:          currentWithSession.Height,
			ActiveSessionID: activeSessionID,
			InUse:           inUse,
		}, nil
	}

	// Perform the update
	_, err = s.loungeRepo.UpdateLoungeStation(ctx, id, updateLabel, updateClosed)
	if err != nil {
		return nil, err
	}

	// Fetch updated station with active session
	updatedWithSession, err := s.loungeRepo.GetLoungeStationByIDWithActiveSession(ctx, id)
	if err != nil {
		return nil, err
	}

	var activeSessionID *string
	inUse := false
	if updatedWithSession.ActiveSessionID.Valid {
		s := updatedWithSession.ActiveSessionID.String()
		activeSessionID = &s
		inUse = true
	}

	return &dto.LoungeStationDTO{
		ID:              updatedWithSession.ID.String(),
		Label:           updatedWithSession.Label,
		StationType:     dto.LoungeStationType(updatedWithSession.StationType),
		Closed:          updatedWithSession.Closed,
		Row:             updatedWithSession.GridRow,
		Col:             updatedWithSession.GridColumn,
		Width:           updatedWithSession.Width,
		Height:          updatedWithSession.Height,
		ActiveSessionID: activeSessionID,
		InUse:           inUse,
	}, nil
}
