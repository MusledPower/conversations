package errs

import "errors"

var (
	ErrInvalidPaginationRequest = errors.New("неверный запрос (некорректные параметры пагинации)")
	ErrInvalidRequest           = errors.New("неверный запрос")
	ErrUnauthorized             = errors.New("не авторизован")
	ErrForbiddenAdmin           = errors.New("доступ запрещён (требуется роль admin)")
	ErrForbidden                = errors.New("доступ запрещён")
	ErrForbiddenUser            = errors.New("доступ запрещён (требуется роль user)")
	ErrInternalServerError      = errors.New("ввнутренняя ошибка сервера")
)

type ErrorResp struct {
	code    string
	message string
}
