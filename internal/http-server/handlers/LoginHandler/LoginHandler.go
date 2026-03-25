package LoginHandler

import (
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
	authService "backend-test/internal/services/auth"
	"backend-test/internal/services/userService"
	"backend-test/internal/storage"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty" validate:"error,omitempty"`
	Token  string `json:"token,omitempty"`
}

// Login godoc
// @Summary Авторизация по email и паролю (ДОПОЛНИТЕЛЬНОЕ ЗАДАНИЕ — необязательно)
// @Description Авторизует пользователя по email и паролю, возвращает JWT.
// @Description Реализация этого эндпоинта является **дополнительным заданием**.
// @Description Для авторизации в рамках обязательной части используйте `/dummyLogin`.
// @Description Доступен без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body Request true "Учётные данные пользователя"
// @Success 200 {object} Response "Успешная авторизация"
// @Failure 401 {object} Response "Неверные учётные данные"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /login [post]
func New(serv *userService.UserService, auth *authService.AuthService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.LoginHandler"

		var req Request

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		err := render.Decode(r, &req)
		fmt.Println(err)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: errs.ErrInvalidRequest.Error()})

			return
		}

		log.Info("request body decoded")

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error(errs.ErrInternalServerError.Error(), sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: errs.ErrInvalidRequest.Error()})

			return
		}

		token, err := serv.AuthUser(auth, req.Email, req.Password)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, Response{Status: http.StatusNotFound, Error: err.Error()})

				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, Response{Status: http.StatusInternalServerError, Error: err.Error()})

			return
		}

		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, Response{Status: http.StatusOK, Token: token})
		return
	}
}
