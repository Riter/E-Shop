package jwt

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var (
	secret    string
	algorithm string
)

type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	AppID  int    `json:"app_id"`
	jwt.RegisteredClaims
}

func init() {
	if err := godotenv.Load("../environment/jwt.env"); err != nil {
		panic("Не удалось загрузить jwt.env: " + err.Error())
	}

	secret = os.Getenv("APP_SECRET")
	if secret == "" {
		panic("APP_SECRET не было получено из jwt.env")
	}

	algorithm = os.Getenv("ALGORITHM")
	if algorithm == "" {
		algorithm = "HS256"
	}
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		//проверка алгоритма
		if token.Method.Alg() != algorithm {
			return nil, fmt.Errorf("неопознанный метод подписи: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid Authorization header format")
	}
	return parts[1], nil
}
