package FindBookingHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/bookingService"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const InvalidRequest = "Неверный запрос"
const InternalServerError = "Внутренняя ошибка сервера"

type BookingListQuery struct {
	Page     int `validate:"gte=1"`
	PageSize int `validate:"gte=1,lte=100"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type Response struct {
	Status     int              `json:"status"`
	Error      string           `json:"error,omitempty" validate:"error,omitempty"`
	Booking    []models.Booking `json:"room,omitempty"`
	Pagination Pagination       `json:"pagination,omitempty"`
}

// MyBookings godoc
// @Summary Общий список броней
// @Description Доступно только роли admin.
//
//	Поддерживает пагинацию через параметры `page` и `pageSize`.
//	Оба параметра опциональны; значения по умолчанию: `page=1`, `pageSize=20`.
//	Максимальное значение `pageSize` — 100.
//
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

		var pageSize int
		var page int

		if r.URL.Query().Get("pageSize") == "" {
			pageSize = 20
		} else {
			pageSize, _ = strconv.Atoi(r.URL.Query().Get("pageSize"))
		}

		if r.URL.Query().Get("page") == "" {
			page = 1
		} else {
			page, _ = strconv.Atoi(r.URL.Query().Get("page"))
		}

		b := BookingListQuery{Page: page, PageSize: pageSize}

		log.Info("request body decoded", slog.Any("request", b))

		if err := validator.New().Struct(b); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error(InvalidRequest, sl.Err(err))

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: InvalidRequest})

			return
		}

		list, total, err := serv.BookingList(b.PageSize, b.Page)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		render.JSON(
			w, r, Response{
				Status:     http.StatusOK,
				Booking:    list,
				Pagination: Pagination{page, pageSize, total},
			},
		)
		return
	}
}
