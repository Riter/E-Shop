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
	
	t.Run("Create Comments with Different Ratings", func(t *testing.T) {
		ratings := []int{5, 4, 3, 2, 1}
		productID := int64(999) 

		for i, rating := range ratings {
			comment := models.CreateCommentDTO{
				ProductID: productID,
				Content:   fmt.Sprintf("Тестовый комментарий с рейтингом %d", rating),
				Rating:    rating,
			}

			jsonData, err := json.Marshal(comment)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "http:
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

		
		t.Run("Check Average Rating", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("http:
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Получение рейтинга продукта - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var rating models.ProductRating
			err = json.NewDecoder(resp.Body).Decode(&rating)
			require.NoError(t, err)

			fmt.Printf("Средний рейтинг: %.2f\n", rating.AverageRating)
			fmt.Printf("Количество отзывов: %d\n", rating.ReviewCount)

			
			expectedAverage := float64(5+4+3+2+1) / 5.0
			require.InDelta(t, expectedAverage, rating.AverageRating, 0.01)
			require.Equal(t, int64(5), rating.ReviewCount)
		})

		
		t.Run("Update Rating", func(t *testing.T) {
			
			updateComment := models.UpdateCommentDTO{
				Content: "Обновленный комментарий с новым рейтингом",
				Rating:  5,
			}

			jsonData, err := json.Marshal(updateComment)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "http:
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "1")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			fmt.Printf("Обновление рейтинга - Статус: %d\n", resp.StatusCode)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			
			resp, err = http.Get(fmt.Sprintf("http:
			require.NoError(t, err)
			defer resp.Body.Close()

			var rating models.ProductRating
			err = json.NewDecoder(resp.Body).Decode(&rating)
			require.NoError(t, err)

			fmt.Printf("Обновленный средний рейтинг: %.2f\n", rating.AverageRating)
			fmt.Printf("Количество отзывов: %d\n", rating.ReviewCount)

			
			expectedAverage := float64(5+4+3+2+1) / 5.0
			require.InDelta(t, expectedAverage, rating.AverageRating, 0.01)
			require.Equal(t, int64(5), rating.ReviewCount)
		})
	})
}
