package roomsRepository

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/storage"
	"backend-test/internal/storage/db"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RoomsRepository struct {
	db *db.Storage
}

func NewRoomsRepository(db *db.Storage) *RoomsRepository {
	return &RoomsRepository{
		db: db,
	}
}

func (r *RoomsRepository) CreateRoom(name string, desc *string, cap *int) (*models.Room, error) {
	const op = "repository.RoomsRepository.CreateRoom"
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	date := time.Now()

	tx, err := r.db.DB.Begin()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO rooms (id, name, description, capacity, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(id, name, desc, cap, date)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Room{
		ID:          id,
		Name:        name,
		Description: desc,
		Capacity:    cap,
		CreatedAt:   date,
	}, nil
}

func (r *RoomsRepository) FindRoomList() ([]models.Room, error) {
	const op = "repository.RoomsRepository.FindRoomList"

	tx, err := r.db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rows, err := tx.Query("SELECT * FROM rooms")

	defer rows.Close()

	if !rows.Next() {
		return nil, storage.ErrRoomNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var rooms []models.Room

	for rows.Next() {
		var room models.Room

		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		rooms = append(rooms, room)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rooms, nil
}
