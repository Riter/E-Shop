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

func TestCommentsService(t *testing.T) {
	// Генерируем токены для разных пользователей
	user1Token, err := generateTestToken(1, "user1@test.com")
	require.NoError(t, err)
	fmt.Printf("Сгенерирован токен для user1: %s\n", user1Token)

	user2Token, err := generateTestToken(2, "user2@test.com")
	require.NoError(t, err)
	fmt.Printf("Сгенерирован токен для user2: %s\n", user2Token)

	// Тест 1: Создание комментария
	t.Run("Create Comment", func(t *testing.T) {
		comment := models.CreateCommentDTO{
			ProductID: 1,
			Content:   "Тестовый комментарий от user1",
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

		fmt.Printf("Создание комментария - Статус: %d\n", resp.StatusCode)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]int64
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.NotZero(t, result["id"])
		fmt.Printf("Создан комментарий с ID: %d\n", result["id"])

		// Сохраняем ID комментария для последующих тестов
		commentID := result["id"]

		// Тест 2: Получение комментариев продукта
		t.Run("Get Product Comments", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/products/1/comments", baseURL))
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Получение комментариев - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var comments []models.Comment
			err = json.NewDecoder(resp.Body).Decode(&comments)
			require.NoError(t, err)
			fmt.Printf("Получено комментариев: %d\n", len(comments))
			for _, c := range comments {
				fmt.Printf("Комментарий ID %d: %s (рейтинг: %d)\n", c.ID, c.Content, c.Rating)
			}
		})

		// Тест 3: Обновление комментария
		t.Run("Update Comment", func(t *testing.T) {
			updateComment := models.UpdateCommentDTO{
				Content: "Обновленный комментарий от user1",
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

			fmt.Printf("Обновление комментария - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// Пробуем обновить чужой комментарий
			req, err = http.NewRequest("PUT", fmt.Sprintf("%s/comments/%d", baseURL, commentID), bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+user2Token)

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Попытка обновить чужой комментарий - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusForbidden, resp.StatusCode)
		})

		// Тест 4: Удаление комментария
		t.Run("Delete Comment", func(t *testing.T) {
			// Пробуем удалить чужой комментарий
			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/comments/%d", baseURL, commentID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+user2Token)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Попытка удалить чужой комментарий - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusForbidden, resp.StatusCode)

			// Удаляем свой комментарий
			req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/comments/%d", baseURL, commentID), nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+user1Token)

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Удаление своего комментария - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	// Тест 5: Проверка невалидных токенов
	t.Run("Invalid Tokens", func(t *testing.T) {
		invalidTokens := []string{
			"",                     // пустой токен
			"invalid-token",        // неверный формат
			"Bearer invalid-token", // неверный формат Bearer
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

			fmt.Printf("Проверка невалидного токена '%s' - Статус: %d\n", token, resp.StatusCode)
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
