package scheduleRepository

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/storage"
	"backend-test/internal/storage/db"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ScheduleRepository struct {
	db *db.Storage
}

func New(db *db.Storage) *ScheduleRepository {
	return &ScheduleRepository{
		db: db,
	}
}

func (r *ScheduleRepository) CreateSchedule(
	roomID uuid.UUID,
	days []int,
	start time.Time,
	end time.Time,
) (*models.Schedule, error) {
	const op = "scheduleRepository.CreateSchedule"

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tx, err := r.db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(id, roomID, days, start, end)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("%w", storage.ErrScheduleExists)
			}

			if pgErr.Code == "23503" {
				return nil, fmt.Errorf("%w", storage.ErrSlotNotFound)
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Schedule{
		ID:         id,
		RoomID:     roomID,
		DaysOfWeek: days,
		StartTime:  start,
		EndTime:    end,
	}, nil
}
