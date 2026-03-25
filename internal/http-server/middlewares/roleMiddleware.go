package middlewares

import (
	"backend-test/internal/lib/errs"
	authService "backend-test/internal/services/auth"
	"net/http"

	"github.com/go-chi/render"
)

type ErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty" validate:"error,omitempty"`
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				claims := r.Context().
					Value("user").(*authService.Token)

				if !hasRole(claims.Role, allowedRoles) {
					render.JSON(
						w, r, &ErrorResponse{
							Status: http.StatusForbidden,
							Error:  errs.ErrForbiddenAdmin.Error(),
						},
					)
					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}

func hasRole(role string, allowedRoles []string) bool {
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}

	return false
}
