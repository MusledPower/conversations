package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Capacity    *int      `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
}
