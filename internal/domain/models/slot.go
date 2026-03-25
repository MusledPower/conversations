package models

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID     uuid.UUID
	RoomID uuid.UUID
	Start  time.Time
	End    time.Time
}
