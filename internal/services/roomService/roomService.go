package roomService

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/repository"
	"fmt"
	"log/slog"
)

type RoomService struct {
	db  repository.RoomsRepository
	log *slog.Logger
}

func NewRoomService(db repository.RoomsRepository, log *slog.Logger) *RoomService {
	return &RoomService{
		db:  db,
		log: log,
	}
}

func (r *RoomService) CreateRoom(name string, capacity *int, desc *string) (*models.Room, error) {
	const op = "services.roomService.createRoom"

	r.log.With("op:", op)

	room, err := r.db.CreateRoom(name, desc, capacity)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	return room, nil
}

func (r *RoomService) FindRoomList() ([]models.Room, error) {
	const op = "services.roomService.findRooms"

	r.log.With("op:", op)

	rooms, err := r.db.FindRoomList()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}

	return rooms, nil
}
