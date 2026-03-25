package DummyLoginHandler

import (
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
	authService "backend-test/internal/services/auth"
	"backend-test/internal/services/userService"
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
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=admin user"`
}

type Response struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty" validate:"error,omitempty"`
	Token  string `json:"token,omitempty"`
}

// DummyLogin godoc
// @Summary Получить тестовый JWT по роли
// @Description Выдаёт тестовый JWT для указанной роли (admin / user). Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param role body object{role=string} true "Роль пользователя (admin/user)"
// @Success 200 {object} Response "JWT токен"
// @Failure 400 {object} Response "Неверный запрос (недопустимая роль)"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/dummyLogin [post]
func New(
	authServ *authService.AuthService,
	userServ *userService.UserService,
	log *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.server.dummyTokenHandler"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.Decode(r, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error(InvalidRequest, sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: errs.ErrInvalidRequest.Error()})

			return
		}

		if err := userServ.FindUser(req.UserID); err != nil {
			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		token, err := authServ.GenerateToken(req.UserID, req.Role)
		if err != nil {
			log.Error(InternalServerError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: InternalServerError})
			return
		}

		render.JSON(w, r, Response{Status: http.StatusOK, Token: token})
		return
	}
}
