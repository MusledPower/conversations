package repository

import (
	"backend-test/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type SlotsRepository interface {
	CreateSlot(roomID uuid.UUID, start time.Time, end time.Time) error
	GetSlotsList(roomID uuid.UUID, date time.Time) ([]models.Slot, error)
}
