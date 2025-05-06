package jwt_test

import (
	"testing"
	"time"

	jwtlocal "comments_service/internal/jwt"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestParseToken_ValidToken(t *testing.T) {
	// Шаг 1: Создаём "сырой" payload с userID, email и app_id
	expectedUserID := int64(42)
	expectedEmail := "test@example.com"
	expectedAppID := 123
	expTime := time.Now().Add(time.Hour)

	claims := jwtlib.MapClaims{
		"uid":    expectedUserID,
		"email":  expectedEmail,
		"app_id": expectedAppID,
		"exp":    jwtlib.NewNumericDate(expTime).Unix(),
		"iss":    "test-suite",
	}

	t.Logf("Создан claims: %+v", claims)

	// Шаг 2: Создаём и подписываем токен
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err, "не удалось подписать токен")

	t.Logf("Подписанный токен: %s", signedToken)

	// Шаг 3: Парсим токен через твою функцию
	parsedClaims, err := jwtlocal.ParseToken(signedToken)
	require.NoError(t, err, "ошибка при парсинге токена")

	t.Logf("Полученные claims: %+v", parsedClaims)

	// Шаг 4: Проверяем userID, email и app_id
	require.Equal(t, expectedUserID, parsedClaims.UserID, "userID не совпадает")
	require.Equal(t, expectedEmail, parsedClaims.Email, "email не совпадает")
	require.Equal(t, expectedAppID, parsedClaims.AppID, "app_id не совпадает")
}
