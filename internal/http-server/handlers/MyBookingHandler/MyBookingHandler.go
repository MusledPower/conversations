package MyBookingHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/bookingService"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const InternalServerError = "Внутренняя ошибка сервера"

type Response struct {
	Status  int              `json:"status"`
	Error   string           `json:"error,omitempty" validate:"error,omitempty"`
	Booking []models.Booking `json:"room,omitempty"`
}

// MyBookings godoc
// @Summary Список броней текущего пользователя
// @Description Возвращает только брони пользователя, чей user_id в JWT. Только будущие слоты.
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} Response "успех"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 403 {object} Response "Доступ запрещён"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/bookings/my [get]
func New(serv *bookingService.BookingService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.MyBookingHandler"

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

		list, err := serv.MyBooking(userID)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		render.JSON(w, r, Response{Status: http.StatusOK, Booking: list})
		return
	}
}
