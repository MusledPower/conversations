package repository

import (
	"backend-test/internal/domain/models"

	"github.com/google/uuid"
)

type BookingRepository interface {
	BookingCreate(userID uuid.UUID, slotID uuid.UUID, conferenceLink string) (*models.Booking, error)
	BookingFind(pageSize int, offset int) ([]models.Booking, int, error)
	BookingMy(userID uuid.UUID) ([]models.Booking, error)
	BookingCancel(userID uuid.UUID, bookingID uuid.UUID) (*models.Booking, error)
}
