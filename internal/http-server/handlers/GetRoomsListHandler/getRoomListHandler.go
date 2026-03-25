package GetRoomsListHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/roomService"
	"backend-test/internal/storage"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	Error string        `json:"error,omitempty"`
	Rooms []models.Room `json:"rooms,omitempty"`
}

type GetRoomListHandler struct {
	serv *roomService.RoomService
	log  *slog.Logger
}

// ListRooms godoc
// @Summary Список переговорок
// @Description Возвращает все переговорки. Доступно для admin и user.
// @Tags Rooms
// @Security BearerAuth
// @Produce json
// @Success 200 {object} Response "Список переговорок"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/rooms/list [get]
func New(serv *roomService.RoomService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.GetRoomsListHandler.getRoomListHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request body decoded")

		rooms, err := serv.FindRoomList()
		if err != nil {
			if errors.Is(err, storage.ErrRoomNotFound) {
				log.Error(storage.ErrRoomNotFound.Error(), sl.Err(err))
			}

			log.Error(errs.ErrInternalServerError.Error(), sl.Err(err))
			render.JSON(w, r, Response{Error: errs.ErrInternalServerError.Error()})
			return
		}

		render.JSON(w, r, Response{Rooms: rooms})
		return
	}
}
