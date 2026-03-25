package middlewares

import (
	"backend-test/internal/lib/errs"
	authService "backend-test/internal/services/auth"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				tokenString := strings.TrimPrefix(
					r.Header.Get("Authorization"),
					"Bearer ",
				)

				if tokenString == "" {
					http.Error(w, "missing token", 401)
					return
				}

				claims := &authService.Token{}

				token, err := jwt.ParseWithClaims(
					tokenString,
					claims,
					func(token *jwt.Token) (interface{}, error) {
						return []byte(secret), nil
					},
				)

				if err != nil || !token.Valid {
					http.Error(w, errs.ErrUnauthorized.Error(), http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
				ctx = context.WithValue(ctx, "user", claims)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			},
		)
	}
}
