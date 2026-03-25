package repository

import "backend-test/internal/domain/models"

type RoomsRepository interface {
	CreateRoom(name string, desc *string, cap *int) (*models.Room, error)
	FindRoomList() ([]models.Room, error)
}
