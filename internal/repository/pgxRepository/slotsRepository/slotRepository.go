package slotsRepository

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/storage/db"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SlotRepository struct {
	db *db.Storage
}

func NewSlotRepository(db *db.Storage) *SlotRepository {
	return &SlotRepository{
		db: db,
	}
}

func (r *SlotRepository) CreateSlot(roomID uuid.UUID, start time.Time, end time.Time) error {
	const op = "pgxRepository.SlotRepository.CreateSlot"

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tx, err := r.db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO slots(id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(id, roomID, start, end)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *SlotRepository) GetSlotsList(roomID uuid.UUID, date time.Time) ([]models.Slot, error) {
	const op = "pgxRepository.SlotRepository.GetSlotsList"
	tx, err := r.db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
        SELECT *
        FROM slots s
        LEFT JOIN bookings b
            ON b.slot_id = s.id
            AND b.status = 'active'
        WHERE s.room_id = $1
          AND DATE(s.start_time) = $2
          AND b.id IS NULL
        ORDER BY s.start_time;
    `

	rows, err := tx.Query(query, roomID, date)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var slots []models.Slot

	for rows.Next() {
		var s models.Slot

		if err := rows.Scan(&s.ID, &roomID, &s.Start, &s.End); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		slots = append(slots, s)
	}

	tx.Commit()

	return slots, nil
}
