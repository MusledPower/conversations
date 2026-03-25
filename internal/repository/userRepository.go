package repository

import (
	"backend-test/internal/domain/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(email string, password string, role models.Role) (*models.User, error)
	FindUser(userID uuid.UUID) error
	FindUserByEmail(email string) (*models.User, error)
}
