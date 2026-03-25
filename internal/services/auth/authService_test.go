package authService

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	secret := "supersecret"
	service := New(secret, 1*time.Hour)

	userID := uuid.New()
	role := "admin"

	tokenString, err := service.GenerateToken(userID, role)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Проверяем, что токен корректный и можно распарсить
	parsedToken, err := jwt.ParseWithClaims(
		tokenString, &Token{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	if claims, ok := parsedToken.Claims.(*Token); ok {
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, role, claims.Role)
		assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, time.Minute)
	} else {
		t.Fatal("Claims are not of type *authService.Token")
	}
}

func TestAuthService_GenerateToken_Error(t *testing.T) {
	// Передаем пустой секрет, что вызовет ошибку при подписи
	service := New("", 24*time.Hour)

	userID := uuid.New()
	role := "admin"

	tokenString, err := service.GenerateToken(userID, role)

	assert.Error(t, err)
	assert.Empty(t, tokenString)
}

func TestAuthService_GenerateToken_InvalidRole(t *testing.T) {
	secret := "supersecret"
	service := New(secret, 24*time.Hour)

	userID := uuid.New()
	role := "guest"

	tokenString, err := service.GenerateToken(userID, role)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
	assert.Empty(t, tokenString)
}
