package slotsService

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/repository"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type SlotsService struct {
	db  repository.SlotsRepository
	log *slog.Logger
}

func New(db repository.SlotsRepository, log *slog.Logger) *SlotsService {
	return &SlotsService{
		db:  db,
		log: log,
	}
}

func (s *SlotsService) CreateSlotsService(roomID uuid.UUID, start time.Time, end time.Time) error {
	const op = "service.slotsService.CreateSlotsService"

	s.log.With("op:", op)

	if err := s.db.CreateSlot(roomID, start, end); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *SlotsService) GetSlotsListService(roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	const op = "service.slotsService.GetSlotsListService"

	s.log.With("op:", op)

	slots, err := s.db.GetSlotsList(roomID, date)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return slots, nil
}
