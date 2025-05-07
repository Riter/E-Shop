package integration

import (
	"bytes"
	"comments_service/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:30333"
)

// generateTestToken создает тестовый JWT токен для указанного пользователя
func generateTestToken(userID int64, email string) (string, error) {
	claims := jwtv5.MapClaims{
		"uid":    userID,
		"email":  email,
		"app_id": 1,
		"exp":    jwtv5.NewNumericDate(time.Now().Add(time.Hour)).Unix(),
		"iss":    "test-suite",
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte("test-secret"))
}

func TestJWTProtectedEndpoints(t *testing.T) {
	// Генерируем токены для разных пользователей
	user1Token, err := generateTestToken(1, "user1@test.com")
	require.NoError(t, err)

	user2Token, err := generateTestToken(2, "user2@test.com")
	require.NoError(t, err)

	// Тест создания комментария
	t.Run("Create Comment", func(t *testing.T) {
		comment := models.CreateCommentDTO{
			ProductID: 1,
			Content:   "Тестовый комментарий",
			Rating:    5,
		}

		jsonData, err := json.Marshal(comment)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/comments", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]int64
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.NotZero(t, result["id"])

		// Сохраняем ID комментария для последующих тестов
		commentID := result["id"]

		// Тест обновления комментария
		t.Run("Update Comment", func(t *testing.T) {
			updateComment := models.UpdateCommentDTO{
				Content: "Обновленный комментарий",
				Rating:  4,
			}

			jsonData, err := json.Marshal(updateComment)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", fmt.Sprintf("%s/comments/%d", baseURL, commentID), bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+user1Token)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			// Пробуем обновить чужой комментарий
			req, err = http.NewRequest("PUT", fmt.Sprintf("%s/comments/%d", baseURL, commentID), bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+user2Token)

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusForbidden, resp.StatusCode)
		})

		// Тест удаления комментария
		t.Run("Delete Comment", func(t *testing.T) {
			// Пробуем удалить чужой комментарий
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/comments/%d", baseURL, commentID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+user2Token)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusForbidden, resp.StatusCode)

			// Удаляем свой комментарий
			req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/comments/%d", baseURL, commentID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+user1Token)

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	// Тест публичных эндпоинтов
	t.Run("Public Endpoints", func(t *testing.T) {
		// Создаем комментарий для тестирования
		comment := models.CreateCommentDTO{
			ProductID: 1,
			Content:   "Публичный тестовый комментарий",
			Rating:    5,
		}

		jsonData, err := json.Marshal(comment)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", baseURL+"/comments", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]int64
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		commentID := result["id"]

		// Тест получения комментария
		t.Run("Get Comment", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/comments/%d", baseURL, commentID))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			var comment models.Comment
			err = json.NewDecoder(resp.Body).Decode(&comment)
			require.NoError(t, err)
			require.Equal(t, commentID, comment.ID)
		})

		// Тест получения комментариев продукта
		t.Run("Get Product Comments", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/products/1/comments", baseURL))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			var comments []models.Comment
			err = json.NewDecoder(resp.Body).Decode(&comments)
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(comments), 1)
		})
	})

	// Тест невалидных токенов
	t.Run("Invalid Tokens", func(t *testing.T) {
		invalidTokens := []string{
			"",                     // пустой токен
			"invalid-token",        // неверный формат
			"Bearer invalid-token", // неверный формат Bearer
			"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", // подписанный другим ключом
		}

		for _, token := range invalidTokens {
			req, err := http.NewRequest("POST", baseURL+"/comments", bytes.NewBufferString("{}"))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			if token != "" {
				req.Header.Set("Authorization", token)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
