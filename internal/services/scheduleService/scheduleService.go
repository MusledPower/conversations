package scheduleService

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/repository"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ScheduleService struct {
	r   repository.ScheduleRepository
	log *slog.Logger
}

func New(r repository.ScheduleRepository, log *slog.Logger) *ScheduleService {
	return &ScheduleService{r: r, log: log}
}

func (s ScheduleService) CreateSchedule(
	roomID uuid.UUID,
	days []int,
	start time.Time,
	end time.Time,
) (*models.Schedule, error) {
	const op = "scheduleService.CreateSchedule"

	s.log.With("op:", op)

	sched, err := s.r.CreateSchedule(roomID, days, start, end)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sched, nil
}
