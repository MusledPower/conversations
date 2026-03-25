package bookingService_test

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/services/bookingService"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок репозитория
type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) BookingCreate(
	userID uuid.UUID,
	slotID uuid.UUID,
	conferenceLink string,
) (*models.Booking, error) {
	args := m.Called(userID, slotID, conferenceLink)
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) BookingFind(pageSize int, page int) ([]models.Booking, int, error) {
	args := m.Called(pageSize, page)
	return args.Get(0).([]models.Booking), args.Int(1), args.Error(2)
}

func (m *MockBookingRepository) BookingMy(userID uuid.UUID) ([]models.Booking, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Booking), args.Error(1)
}

func (m *MockBookingRepository) BookingCancel(userID uuid.UUID, bookingID uuid.UUID) (*models.Booking, error) {
	args := m.Called(userID, bookingID)
	return args.Get(0).(*models.Booking), args.Error(1)
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestBookingList(t *testing.T) {
	fixedTime := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)

	mockBookings := []models.Booking{
		{
			ID:        uuid.New(),
			SlotID:    uuid.New(),
			UserID:    uuid.New(),
			Status:    "active",
			CreatedAt: fixedTime,
		},
		{
			ID:        uuid.New(),
			SlotID:    uuid.New(),
			UserID:    uuid.New(),
			Status:    "cancelled",
			CreatedAt: fixedTime,
		},
	}

	tests := []struct {
		name          string
		pageSize      int
		page          int
		mockBookings  []models.Booking
		mockTotal     int
		mockErr       error
		expectedList  []models.Booking
		expectedTotal int
		expectedErr   bool
	}{
		{
			name:          "успешное получение списка",
			pageSize:      20,
			page:          1,
			mockBookings:  mockBookings,
			mockTotal:     2,
			mockErr:       nil,
			expectedList:  mockBookings,
			expectedTotal: 2,
			expectedErr:   false,
		},
		{
			name:          "пустой список",
			pageSize:      20,
			page:          1,
			mockBookings:  []models.Booking{},
			mockTotal:     0,
			mockErr:       nil,
			expectedList:  []models.Booking{},
			expectedTotal: 0,
			expectedErr:   false,
		},
		{
			name:          "вторая страница",
			pageSize:      1,
			page:          2,
			mockBookings:  mockBookings[1:],
			mockTotal:     2,
			mockErr:       nil,
			expectedList:  mockBookings[1:],
			expectedTotal: 2,
			expectedErr:   false,
		},
		{
			name:          "ошибка репозитория",
			pageSize:      20,
			page:          1,
			mockBookings:  nil,
			mockTotal:     0,
			mockErr:       errors.New("database error"),
			expectedList:  nil,
			expectedTotal: 0,
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mockRepo := new(MockBookingRepository)
				mockRepo.On("BookingFind", tt.pageSize, tt.page).
					Return(tt.mockBookings, tt.mockTotal, tt.mockErr)

				svc := bookingService.New(mockRepo, newLogger())

				list, total, err := svc.BookingList(tt.pageSize, tt.page)

				if tt.expectedErr {
					assert.Error(t, err)
					assert.Nil(t, list)
					assert.Equal(t, 0, total)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedList, list)
					assert.Equal(t, tt.expectedTotal, total)
				}

				mockRepo.AssertExpectations(t)
			},
		)
	}
}
