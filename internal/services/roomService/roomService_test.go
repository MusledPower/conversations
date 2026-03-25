package roomService_test

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/services/roomService"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Создаем мок для RoomsRepository
type MockRoomsRepository struct {
	mock.Mock
}

func (m *MockRoomsRepository) CreateRoom(name string, desc *string, capacity *int) (*models.Room, error) {
	args := m.Called(name, desc, capacity)
	if room := args.Get(0); room != nil {
		return room.(*models.Room), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoomsRepository) FindRoomList() ([]models.Room, error) {
	args := m.Called()
	if list := args.Get(0); list != nil {
		return list.([]models.Room), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestRoomService_CreateRoom_Success(t *testing.T) {
	mockRepo := new(MockRoomsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	capacity := 10
	desc := "Test room"
	expectedRoom := &models.Room{
		Name:        "Room1",
		Description: &desc,
		Capacity:    &capacity,
	}

	mockRepo.On("CreateRoom", "Room1", &desc, &capacity).Return(expectedRoom, nil)

	service := roomService.NewRoomService(mockRepo, logger)
	room, err := service.CreateRoom("Room1", &capacity, &desc)

	assert.NoError(t, err)
	assert.Equal(t, expectedRoom, room)
	mockRepo.AssertExpectations(t)
}

func TestRoomService_CreateRoom_Error(t *testing.T) {
	mockRepo := new(MockRoomsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	capacity := 10
	desc := "Test room"
	expectedErr := errors.New("db error")

	mockRepo.On("CreateRoom", "Room1", &desc, &capacity).Return(nil, expectedErr)

	service := roomService.NewRoomService(mockRepo, logger)
	room, err := service.CreateRoom("Room1", &capacity, &desc)

	assert.Nil(t, room)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}

func TestRoomService_FindRoomList_Success(t *testing.T) {
	mockRepo := new(MockRoomsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	expectedRooms := []models.Room{
		{Name: "Room1"},
		{Name: "Room2"},
	}

	mockRepo.On("FindRoomList").Return(expectedRooms, nil)

	service := roomService.NewRoomService(mockRepo, logger)
	rooms, err := service.FindRoomList()

	assert.NoError(t, err)
	assert.Equal(t, expectedRooms, rooms)
	mockRepo.AssertExpectations(t)
}

func TestRoomService_FindRoomList_Error(t *testing.T) {
	mockRepo := new(MockRoomsRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	expectedErr := errors.New("db error")
	mockRepo.On("FindRoomList").Return(nil, expectedErr)

	service := roomService.NewRoomService(mockRepo, logger)
	rooms, err := service.FindRoomList()

	assert.Nil(t, rooms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	mockRepo.AssertExpectations(t)
}
