package bookingRepository

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/storage"
	"backend-test/internal/storage/db"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type BookingRepository struct {
	db *db.Storage
}

func New(db *db.Storage) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) BookingCreate(userID uuid.UUID, slotID uuid.UUID, conf string) (*models.Booking, error) {
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

	query := `
			INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
			SELECT $1, $2, $3, $4, $5, $6
			FROM slots s
			WHERE s.id = $2
  			AND s.start_time > now();`

	stmt, err := tx.Prepare(query)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	cmdTag, err := stmt.Exec(
		id,
		slotID,
		userID,
		"active",
		url.QueryEscape(conf),
		date,
	)
	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("%w", storage.ErrSlotAlreadyBooked)
			}

			if pgErr.Code == "23503" {
				return nil, fmt.Errorf("%w", storage.ErrSlotNotFound)
			}
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := cmdTag.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if rows == 0 {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.Booking{
		ID:             id,
		SlotID:         slotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: conf,
		CreatedAt:      date,
	}, nil
}

func (r *BookingRepository) BookingFind(pageSize int, page int) ([]models.Booking, int, error) {
	const op = "pgxRepository.BookingRepository.BookingFind"

	offset := (page - 1) * pageSize

	var total int

	tx, err := r.db.DB.Begin()
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = r.db.DB.QueryRow(`SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := r.db.DB.Query(
		`
		SELECT 
		    id,
		    slot_id,
		    user_id,
		    status,
		    conference_link,
		    created_at
		FROM bookings
		ORDER BY created_at
		LIMIT $1 OFFSET $2`,
		pageSize,
		offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var bookings []models.Booking

	for rows.Next() {
		var b models.Booking

		err := rows.Scan(
			&b.ID,
			&b.SlotID,
			&b.UserID,
			&b.Status,
			&b.ConferenceLink,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}

		bookings = append(bookings, b)
	}

	return bookings, total, nil
}

func (r *BookingRepository) BookingMy(userID uuid.UUID) ([]models.Booking, error) {
	const op = "pgxRepository.BookingRepository.BookingMy"

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
	SELECT
	    b.id AS booking_id,
	    b.slot_id,
	    b.user_id,
	    b.status,
	    b.conference_link,
	    b.created_at
	FROM bookings b
	INNER JOIN slots s ON b.slot_id = s.id
	WHERE b.user_id = $1
	  AND s.start_time >= NOW()
	ORDER BY s.start_time;
	`

	rows, err := tx.Query(query, userID)

	var bookings []models.Booking
	defer rows.Close()

	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(
			&b.ID,
			&b.SlotID,
			&b.UserID,
			&b.Status,
			&b.ConferenceLink,
			&b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		bookings = append(bookings, b)
	}

	tx.Commit()

	return bookings, nil
}

func (r *BookingRepository) BookingCancel(userID uuid.UUID, bookingID uuid.UUID) (*models.Booking, error) {
	const op = "pgxRepository.BookingRepository.BookingMy"

	tx, err := r.db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var ownerID uuid.UUID
	err = tx.QueryRow(
		`SELECT user_id FROM bookings WHERE id = $1`,
		bookingID,
	).Scan(&ownerID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", storage.ErrBookingNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if ownerID != userID {
		return nil, fmt.Errorf("%w", storage.ErrBookingForbidden)
	}

	var booking models.Booking

	err = tx.QueryRow(
		`UPDATE bookings SET status = 'cancelled' 
         WHERE id = $1 
         RETURNING id, user_id, slot_id, status, created_at`,
		bookingID,
	).Scan(&booking.ID, &booking.UserID, &booking.SlotID, &booking.Status, &booking.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &booking, nil
}
