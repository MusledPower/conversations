package userService

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/repository"
	authService "backend-test/internal/services/auth"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repository.UserRepository
	log  *slog.Logger
}

func New(repo repository.UserRepository, log *slog.Logger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

func (s *UserService) CreateUser(email string, password string, secondPassword string) (*models.User, error) {
	const op = "userService.CreateUser"

	s.log.With("op", op)

	// Сначала проверяем совпадение паролей в открытом виде
	if password != secondPassword {
		return nil, fmt.Errorf("%s: passwords do not match", op)
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаём пользователя
	user, err := s.repo.CreateUser(email, hashedPassword, "user")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *UserService) FindUser(userID uuid.UUID) error {
	const op = "userService.FindUser"

	s.log.With("op:", op)

	err := s.repo.FindUser(userID)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	return nil
}

func (s *UserService) AuthUser(authServ *authService.AuthService, email string, password string) (string, error) {
	const op = "userService.AuthUser"

	user, err := s.repo.FindUserByEmail(email)
	fmt.Println("11111")
	fmt.Println(user.Password)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	token, err := authServ.GenerateToken(user.ID, string(user.Role))
	return token, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(hash), nil
}
