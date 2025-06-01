package integration

import (
	"comments_service/internal/db"
	"comments_service/internal/models"
	"comments_service/internal/repository"
	"comments_service/internal/service"
	"fmt"
	"testing"
)

func TestCommentCRUD(t *testing.T) {
	
	db, err := db.InitPsqlDB()
	if err != nil {
		t.Fatalf("ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	
	repo := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(repo, db)

	
	t.Run("Create Comment", func(t *testing.T) {
		createComment := models.CreateCommentDTO{
			UserID:    1,
			ProductID: 1,
			Content:   "Отличный товар! Очень доволен покупкой.",
			Rating:    5,
		}

		commentID, err := commentService.CreateComment(createComment)
		if err != nil {
			t.Fatalf("ошибка создания комментария: %v", err)
		}
		t.Logf("Создан комментарий с ID: %d", commentID)

		
		comment, err := commentService.GetComment(commentID)
		if err != nil {
			t.Fatalf("ошибка получения комментария: %v", err)
		}
		if comment == nil {
			t.Fatal("комментарий не найден после создания")
		}
		if comment.Content != createComment.Content {
			t.Errorf("ожидался контент %s, получен %s", createComment.Content, comment.Content)
		}
	})

	
	t.Run("Get Product Comments", func(t *testing.T) {
		
		for i := 1; i <= 3; i++ {
			createComment := models.CreateCommentDTO{
				UserID:    int64(i),
				ProductID: 1,
				Content:   fmt.Sprintf("Тестовый комментарий #%d", i),
				Rating:    i + 2,
			}
			_, err := commentService.CreateComment(createComment)
			if err != nil {
				t.Fatalf("ошибка создания тестового комментария #%d: %v", i, err)
			}
		}

		
		comments, err := commentService.GetProductComments(1)
		if err != nil {
			t.Fatalf("ошибка получения комментариев продукта: %v", err)
		}

		
		if len(comments) < 3 {
			t.Errorf("ожидалось минимум 3 комментария, получено %d", len(comments))
		}
	})

	
	t.Run("Update Comment", func(t *testing.T) {
		
		createComment := models.CreateCommentDTO{
			UserID:    1,
			ProductID: 1,
			Content:   "Исходный комментарий",
			Rating:    3,
		}
		commentID, err := commentService.CreateComment(createComment)
		if err != nil {
			t.Fatalf("ошибка создания комментария для обновления: %v", err)
		}

		
		updateComment := models.UpdateCommentDTO{
			Content: "Обновленный комментарий",
			Rating:  5,
		}
		err = commentService.UpdateComment(commentID, updateComment)
		if err != nil {
			t.Fatalf("ошибка обновления комментария: %v", err)
		}

		
		updatedComment, err := commentService.GetComment(commentID)
		if err != nil {
			t.Fatalf("ошибка получения обновленного комментария: %v", err)
		}
		if updatedComment.Content != updateComment.Content {
			t.Errorf("ожидался контент %s, получен %s", updateComment.Content, updatedComment.Content)
		}
		if updatedComment.Rating != updateComment.Rating {
			t.Errorf("ожидался рейтинг %d, получен %d", updateComment.Rating, updatedComment.Rating)
		}
	})

	
	t.Run("Delete Comment", func(t *testing.T) {
		
		createComment := models.CreateCommentDTO{
			UserID:    1,
			ProductID: 1,
			Content:   "Комментарий для удаления",
			Rating:    4,
		}
		commentID, err := commentService.CreateComment(createComment)
		if err != nil {
			t.Fatalf("ошибка создания комментария для удаления: %v", err)
		}

		
		err = commentService.DeleteComment(commentID)
		if err != nil {
			t.Fatalf("ошибка удаления комментария: %v", err)
		}

		
		deletedComment, err := commentService.GetComment(commentID)
		if err != nil {
			t.Fatalf("ошибка при проверке удаления: %v", err)
		}
		if deletedComment != nil {
			t.Error("комментарий все еще существует после удаления")
		}
	})
}
