package authService

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Token struct {
	UserID uuid.UUID
	Role   string
	jwt.RegisteredClaims
}

type AuthService struct {
	jwtSecret      string
	expirationTime time.Duration
}

func New(jwtSecret string, expirationTime time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:      jwtSecret,
		expirationTime: expirationTime,
	}
}

func (a *AuthService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	const op = "service.AuthService.GenerateToken"

	claims := Token{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}
