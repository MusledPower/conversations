package BookingCreateHandler

import (
	"backend-test/internal/domain/models"
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
	SlotID         uuid.UUID `json:"slot_id" validate:"required,uuid"`
	ConferenceLink string    `json:"conference_link"`
}

type Response struct {
	Status  int             `json:"status"`
	Error   string          `json:"error,omitempty" validate:"error,omitempty"`
	Booking *models.Booking `json:"room,omitempty"`
}

// CreateBooking godoc
// @Summary Создать бронь на слот
// @Description Доступно только роли user. userId берётся из JWT. Можно опционально запросить конференцию.
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param booking body object{slotId=string,createConferenceLink=bool} true "Данные брони"
// @Success 201 {object} Response "Бронь создана"
// @Failure 400 {object} Response "Неверный запрос"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 403 {object} Response "Доступ запрещён"
// @Failure 404 {object} Response "Слот не найден"
// @Failure 409 {object} Response "Слот уже занят"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/bookings/create [post]
func New(serv *bookingService.BookingService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.BookingCreateHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		userID, ok := r.Context().Value("user_id").(uuid.UUID)
		if !ok {
			log.Error("user id not found in context")

			w.WriteHeader(http.StatusUnauthorized)

			render.JSON(
				w, r, Response{
					Status: http.StatusUnauthorized,
					Error:  "unauthorized",
				},
			)

			return
		}

		err := render.Decode(r, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(
				w, r, Response{
					Status: http.StatusBadRequest,
					Error:  InvalidRequest,
				},
			)

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

		booking, err := serv.CreateBooking(userID, req.SlotID, req.ConferenceLink)
		if err != nil {
			if errors.Is(err, storage.ErrSlotNotFound) {
				log.Error(storage.ErrSlotNotFound.Error(), sl.Err(err))
				w.WriteHeader(http.StatusNotFound)

				render.JSON(
					w, r, Response{
						Status: http.StatusNotFound,
						Error:  storage.ErrSlotNotFound.Error(),
					},
				)
				return
			}

			if errors.Is(err, storage.ErrSlotAlreadyBooked) {
				log.Error(storage.ErrSlotNotFound.Error(), sl.Err(err))
				w.WriteHeader(http.StatusConflict)

				render.JSON(
					w, r, Response{
						Status: http.StatusConflict,
						Error:  storage.ErrSlotAlreadyBooked.Error(),
					},
				)

				return
			}

			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusNotFound)

			render.JSON(
				w, r, Response{
					Status: http.StatusInternalServerError,
					Error:  InternalServerError,
				},
			)

			return
		}

		w.WriteHeader(http.StatusCreated)

		render.JSON(
			w, r, Response{
				Status:  http.StatusCreated,
				Booking: booking,
			},
		)

		return
	}
}
