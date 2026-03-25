package usersRepository

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/storage"
	"backend-test/internal/storage/db"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type UsersRepository struct {
	db *db.Storage
}

func New(db *db.Storage) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) CreateUser(email string, password string, role models.Role) (*models.User, error) {
	const op = "pgxRepository.BookingRepository.CreateBooking"

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

	stmt, err := tx.Prepare("INSERT INTO users (id, email, password, role, created_at) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(
		id,
		email,
		password,
		role,
		date,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrEmailAlreadyExist)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.User{
		ID:        id,
		Email:     email,
		Role:      role,
		CreatedAt: date,
	}, nil
}

func (r *UsersRepository) FindUser(userID uuid.UUID) error {
	const op = "pgxRepository.BookingRepository.FindUser"

	tx, err := r.db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	rows, err := tx.Query("SELECT * FROM users WHERE id = $1", userID)

	defer rows.Close()

	if !rows.Next() {
		return storage.ErrUserNotFound
	}

	return nil
}
func (r *UsersRepository) FindUserByEmail(email string) (*models.User, error) {
	const op = "pgxRepository.BookingRepository.FindUser"

	var user models.User

	query := `
        SELECT id, email, password, role, created_at
        FROM users
        WHERE email = $1
    `

	row := r.db.DB.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
