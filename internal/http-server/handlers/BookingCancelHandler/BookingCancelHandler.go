package BookingCancelHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/bookingService"
	"backend-test/internal/storage"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const InvalidRequest = "Неверный запрос"
const InternalServerError = "Внутренняя ошибка сервера"

type Request struct {
	BookingID uuid.UUID `json:"booking_id" validate:"required, uuid"`
}

type Response struct {
	Status  int             `json:"status"`
	Error   string          `json:"error,omitempty" validate:"error,omitempty"`
	Booking *models.Booking `json:"room,omitempty"`
}

// CancelBooking godoc
// @Summary Отменить бронь
// @Description Доступно только роли user. Можно отменять только свои брони. Идемпотентно.
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param bookingId path string true "UUID брони"
// @Success 200 {object} Response "Бронь отменена"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 403 {object} Response "Не своя бронь или роль не user"
// @Failure 404 {object} Response "Бронь не найдена"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/bookings/{bookingId}/cancel [post]
func New(serv *bookingService.BookingService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.BookingCancelHandler.BookingCancel"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, ok := r.Context().Value("user_id").(uuid.UUID)
		if !ok {
			log.Error("user id not found in context")

			w.WriteHeader(http.StatusUnauthorized)

			render.JSON(
				w, r, Response{
					Status: http.StatusUnauthorized,
					Error:  errs.ErrUnauthorized.Error(),
				},
			)

			return
		}

		var req Request

		err := render.Decode(r, &req)

		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: InvalidRequest})

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error(InvalidRequest, sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(
				w, r, Response{
					Status: http.StatusBadRequest,
					Error:  InvalidRequest,
				},
			)

			return
		}

		booking, err := serv.CancelBooking(req.BookingID, userID)
		if err != nil {
			if errors.Is(err, storage.ErrBookingNotFound) {
				log.Error(storage.ErrBookingNotFound.Error(), sl.Err(err))
				w.WriteHeader(http.StatusNotFound)

				render.JSON(
					w, r, Response{
						Status: http.StatusNotFound,
						Error:  storage.ErrSlotNotFound.Error(),
					},
				)
				return
			}

			if errors.Is(err, storage.ErrBookingForbidden) {
				log.Error(storage.ErrBookingForbidden.Error(), sl.Err(err))
				w.WriteHeader(http.StatusForbidden)

				render.JSON(
					w, r, Response{
						Status: http.StatusForbidden,
						Error:  storage.ErrBookingForbidden.Error(),
					},
				)
				return
			}

			log.Error(InternalServerError, sl.Err(err))
			render.JSON(
				w, r, Response{
					Status: http.StatusInternalServerError,
					Error:  InternalServerError,
				},
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(
			w, r, Response{
				Status:  http.StatusOK,
				Booking: booking,
			},
		)
		return
	}
}
