package GetSlotsHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/slotsService"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const InternalServerError = "Внутренняя ошибка сервера"

type Response struct {
	Status int           `json:"status"`
	Error  string        `json:"error,omitempty"`
	Slots  []models.Slot `json:"slots,omitempty"`
}

// ListSlots godoc
// @Summary Список доступных слотов для переговорки
// @Description Возвращает слоты на указанную дату, которые ещё не забронированы. Доступно для admin и user.
// @Tags Slots
// @Security BearerAuth
// @Produce json
// @Param roomId path string true "UUID переговорки"
// @Param date query string true "Дата в формате YYYY-MM-DD"
// @Success 200 {object} Response "Список доступных слотов"
// @Failure 400 {object} Response "Неверный запрос (отсутствует или некорректен параметр date)"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 404 {object} Response "Переговорка не найдена"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/rooms/{roomId}/slots/list [get]
func New(serv *slotsService.SlotsService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetSlotsHandler.getSlotsHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request body decoded")

		roomID, err := uuid.Parse(chi.URLParam(r, "roomId"))
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		date, err := time.Parse("2006-01-02", r.URL.Query().Get("date"))
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		slots, err := serv.GetSlotsListService(roomID, date)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		render.JSON(w, r, Response{Status: http.StatusCreated, Slots: slots})
		return
	}

}
