package createRoomHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/sl"
	"backend-test/internal/services/roomService"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const InvalidRequest = "Неверный запрос"
const InternalServerError = "Внутренняя ошибка сервера"

type Request struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty"`
	Capacity    *int    `json:"capacity,omitempty"`
}

type Response struct {
	Status int          `json:"status"`
	Error  string       `json:"error,omitempty" validate:"error,omitempty"`
	Room   *models.Room `json:"room,omitempty"`
}

// CreateRoom godoc
// @Summary Создать переговорку
// @Description Доступно только роли admin.
// @Tags Rooms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param room body object{name=string,description=string,capacity=int} true "Данные переговорки"
// @Success 201 {object} Response "Переговорка создана"
// @Failure 400 {object} Response "Неверный запрос"
// @Failure 401 {object} Response "Не авторизован"
// @Failure 403 {object} Response "Доступ запрещён"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/rooms/create [post]
func New(
	serv *roomService.RoomService,
	log *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.roomHandler.RoomHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.Decode(r, &req)
		if errors.Is(err, io.EOF) {
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

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: InvalidRequest})

			return
		}

		room, err := serv.CreateRoom(req.Name, req.Capacity, req.Description)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		render.JSON(w, r, Response{Status: http.StatusCreated, Room: room})
		return
	}
}
