package slotsService

import (
	"backend-test/internal/domain/models"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для SlotsRepository
type MockSlotsRepository struct {
	mock.Mock
}

func (m *MockSlotsRepository) CreateSlot(roomID uuid.UUID, start, end time.Time) error {
	args := m.Called(roomID, start, end)
	return args.Error(0)
}

func (m *MockSlotsRepository) GetSlotsList(roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	args := m.Called(roomID, date)
	if slots := args.Get(0); slots != nil {
		return slots.([]models.Slot), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestSlotsService_CreateSlotsService_Success(t *testing.T) {
	mockRepo := new(MockSlotsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	start := time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 24, 17, 0, 0, 0, time.UTC)

	mockRepo.On("CreateSlot", roomID, start, end).Return(nil)

	service := New(mockRepo, logger)
	err := service.CreateSlotsService(roomID, start, end)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSlotsService_CreateSlotsService_Error(t *testing.T) {
	mockRepo := new(MockSlotsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	start := time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 24, 17, 0, 0, 0, time.UTC)
	expectedErr := errors.New("db error")

	mockRepo.On("CreateSlot", roomID, start, end).Return(expectedErr)

	service := New(mockRepo, logger)
	err := service.CreateSlotsService(roomID, start, end)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}

func TestSlotsService_GetSlotsListService_Success(t *testing.T) {
	mockRepo := new(MockSlotsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	date := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	expectedSlots := []models.Slot{
		{ID: uuid.New(), Start: date.Add(9 * time.Hour), End: date.Add(10 * time.Hour)},
		{ID: uuid.New(), Start: date.Add(10 * time.Hour), End: date.Add(11 * time.Hour)},
	}

	mockRepo.On("GetSlotsList", roomID, date).Return(expectedSlots, nil)

	service := New(mockRepo, logger)
	slots, err := service.GetSlotsListService(roomID, date)

	assert.NoError(t, err)
	assert.Equal(t, expectedSlots, slots)
	mockRepo.AssertExpectations(t)
}

func TestSlotsService_GetSlotsListService_Error(t *testing.T) {
	mockRepo := new(MockSlotsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	date := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("db error")

	mockRepo.On("GetSlotsList", roomID, date).Return(nil, expectedErr)

	service := New(mockRepo, logger)
	slots, err := service.GetSlotsListService(roomID, date)

	assert.Nil(t, slots)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}
