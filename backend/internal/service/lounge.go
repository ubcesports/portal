package service

import (
	"context"

	"github.com/ubcesports/memberships/internal/dto"
	"github.com/ubcesports/memberships/internal/repository"
)

type LoungeService struct {
	loungeRepo *repository.LoungeRepository
}

func NewLoungeService(loungeRepo *repository.LoungeRepository) *LoungeService {
	return &LoungeService{loungeRepo: loungeRepo}
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
