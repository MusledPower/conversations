package scheduleService

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

// Мок для ScheduleRepository
type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) CreateSchedule(roomID uuid.UUID, days []int, start, end time.Time) (
	*models.Schedule,
	error,
) {
	args := m.Called(roomID, days, start, end)
	if sched := args.Get(0); sched != nil {
		return sched.(*models.Schedule), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestScheduleService_CreateSchedule_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	days := []int{1, 3, 5}
	start := time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 24, 17, 0, 0, 0, time.UTC)
	expectedSchedule := &models.Schedule{
		RoomID:     roomID,
		DaysOfWeek: days,
		StartTime:  start,
		EndTime:    end,
	}

	mockRepo.On("CreateSchedule", roomID, days, start, end).Return(expectedSchedule, nil)

	service := New(mockRepo, logger)
	sched, err := service.CreateSchedule(roomID, days, start, end)

	assert.NoError(t, err)
	assert.Equal(t, expectedSchedule, sched)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_CreateSchedule_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	roomID := uuid.New()
	days := []int{1, 3, 5}
	start := time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 24, 17, 0, 0, 0, time.UTC)
	expectedErr := errors.New("db error")

	mockRepo.On("CreateSchedule", roomID, days, start, end).Return(nil, expectedErr)

	service := New(mockRepo, logger)
	sched, err := service.CreateSchedule(roomID, days, start, end)

	assert.Nil(t, sched)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}
