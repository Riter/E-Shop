package integration

import (
	"bytes"
	"comments_service/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:30333"
)

func TestCommentsService(t *testing.T) {
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
		req.Header.Set("X-User-ID", "1")

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
			req.Header.Set("X-User-ID", "1")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Обновление комментария - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// Пробуем обновить чужой комментарий
			req, err = http.NewRequest("PUT", fmt.Sprintf("%s/comments/%d", baseURL, commentID), bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "2")

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
			req.Header.Set("X-User-ID", "2")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Попытка удалить чужой комментарий - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusForbidden, resp.StatusCode)

			// Удаляем свой комментарий
			req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/comments/%d", baseURL, commentID), nil)
			require.NoError(t, err)
			req.Header.Set("X-User-ID", "1")

			resp, err = client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Удаление своего комментария - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	// Тест 5: Проверка отсутствия заголовка X-User-ID
	t.Run("Missing X-User-ID", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURL+"/comments", bytes.NewBufferString("{}"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		fmt.Printf("Проверка отсутствия X-User-ID - Статус: %d\n", resp.StatusCode)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
