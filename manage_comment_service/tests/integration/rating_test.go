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

func TestProductRating(t *testing.T) {
	// Тест 1: Создание комментариев с разными рейтингами
	t.Run("Create Comments with Different Ratings", func(t *testing.T) {
		ratings := []int{5, 4, 3, 2, 1}
		productID := int64(999) // Используем уникальный ID продукта для тестов

		for i, rating := range ratings {
			comment := models.CreateCommentDTO{
				ProductID: productID,
				Content:   fmt.Sprintf("Тестовый комментарий с рейтингом %d", rating),
				Rating:    rating,
			}

			jsonData, err := json.Marshal(comment)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "http://localhost:30333/comments", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", i+1))

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Создание комментария с рейтингом %d - Статус: %d\n", rating, resp.StatusCode)
			require.Equal(t, http.StatusCreated, resp.StatusCode)
		}

		// Тест 2: Проверка среднего рейтинга
		t.Run("Check Average Rating", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("http://localhost:30333/products/%d/rating", productID))
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Получение рейтинга продукта - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var rating models.ProductRating
			err = json.NewDecoder(resp.Body).Decode(&rating)
			require.NoError(t, err)

			fmt.Printf("Средний рейтинг: %.2f\n", rating.AverageRating)
			fmt.Printf("Количество отзывов: %d\n", rating.ReviewCount)

			// Проверяем, что средний рейтинг соответствует ожидаемому
			expectedAverage := float64(5+4+3+2+1) / 5.0
			require.InDelta(t, expectedAverage, rating.AverageRating, 0.01)
			require.Equal(t, int64(5), rating.ReviewCount)
		})

		// Тест 3: Обновление рейтинга
		t.Run("Update Rating", func(t *testing.T) {
			// Обновляем рейтинг первого комментария
			updateComment := models.UpdateCommentDTO{
				Content: "Обновленный комментарий с новым рейтингом",
				Rating:  5,
			}

			jsonData, err := json.Marshal(updateComment)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "http://localhost:30333/comments/1", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "1")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Обновление рейтинга - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// Проверяем обновленный средний рейтинг
			resp, err = http.Get(fmt.Sprintf("http://localhost:30333/products/%d/rating", productID))
			require.NoError(t, err)
			defer resp.Body.Close()

			var rating models.ProductRating
			err = json.NewDecoder(resp.Body).Decode(&rating)
			require.NoError(t, err)

			fmt.Printf("Обновленный средний рейтинг: %.2f\n", rating.AverageRating)
			fmt.Printf("Количество отзывов: %d\n", rating.ReviewCount)

			// Проверяем, что средний рейтинг обновился
			expectedAverage := float64(5+4+3+2+1) / 5.0
			require.InDelta(t, expectedAverage, rating.AverageRating, 0.01)
			require.Equal(t, int64(5), rating.ReviewCount)
		})
	})
}
