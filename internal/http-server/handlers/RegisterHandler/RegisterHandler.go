package RegisterHandler

import (
	"backend-test/internal/domain/models"
	"backend-test/internal/lib/errs"
	"backend-test/internal/lib/sl"
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
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	SecondPassword string `json:"second_password" validate:"eqfield=Password"`
}

type Response struct {
	Status int          `json:"status"`
	Error  string       `json:"error,omitempty" validate:"error,omitempty"`
	User   *models.User `json:"user,omitempty"`
}

// RegisterUser godoc
// @Summary Создать пользователя
// @Description Создаёт нового пользователя (доп. задание). Доступно без авторизации.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body object{email=string,password=string,role=string} true "Данные нового пользователя"
// @Success 201 {object} Response "Пользователь создан"
// @Failure 400 {object} Response "Неверный запрос"
// @Failure 500 {object} Response "Внутренняя ошибка сервера"
// @Router /api/register [post]
func New(serv *userService.UserService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.Register.New"

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

		user, err := serv.CreateUser(req.Email, req.Password, req.SecondPassword)
		if err != nil {
			if errors.Is(err, storage.ErrEmailAlreadyExist) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: err.Error()})

				return
			}

			w.WriteHeader(http.StatusBadRequest)
			log.Error(errs.ErrInternalServerError.Error(), sl.Err(err))
			render.JSON(w, r, Response{Status: http.StatusBadRequest, Error: errs.ErrInvalidRequest.Error()})

			return
		}
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, Response{Status: http.StatusCreated, User: user})

		return
	}
}
