package repository

import (
	"backend-test/internal/domain/models"
	"time"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	CreateSchedule(
		roomID uuid.UUID,
		days []int,
		start time.Time,
		end time.Time,
	) (*models.Schedule, error)
}
