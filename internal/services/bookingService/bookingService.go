package bookingService

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/repository"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type BookingService struct {
	repo repository.BookingRepository
	log  *slog.Logger
}

func New(repo repository.BookingRepository, log *slog.Logger) *BookingService {
	return &BookingService{
		repo: repo,
		log:  log,
	}
}

func (b *BookingService) CreateBooking(
	userID uuid.UUID,
	slotID uuid.UUID,
	conferenceLink string,
) (
	*models.Booking,
	error,
) {
	const op = "services.booking.createBooking"

	b.log.With("op", op)

	booking, err := b.repo.BookingCreate(userID, slotID, conferenceLink)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}

func (b *BookingService) BookingList(pageSize, page int) ([]models.Booking, int, error) {
	const op = "services.booking.bookingList"

	b.log.With("op", op)

	list, total, err := b.repo.BookingFind(pageSize, page)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return list, total, nil
}

func (b *BookingService) MyBooking(userID uuid.UUID) ([]models.Booking, error) {
	const op = "services.booking.MyBooking"

	b.log.With("op", op)

	list, err := b.repo.BookingMy(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return list, nil
}

func (b *BookingService) CancelBooking(bookingID uuid.UUID, userID uuid.UUID) (*models.Booking, error) {
	const op = "services.booking.CancelBooking"

	b.log.With("op", op)

	booking, err := b.repo.BookingCancel(userID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return booking, nil
}
