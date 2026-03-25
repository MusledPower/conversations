package models

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek []int
	StartTime  time.Time
	EndTime    time.Time
}
