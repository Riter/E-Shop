package handler

import (
	"comments_service/internal/auth"
	"comments_service/internal/jwt"
	"comments_service/internal/models"
	"comments_service/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// contextKey - тип для ключей контекста
type contextKey string

const (
	// userIDKey - ключ для хранения ID пользователя в контексте
	userIDKey contextKey = "userID"
)

type CommentHandler struct {
	service   *service.CommentService
	ssoClient *auth.Client
}

func NewCommentHandler(service *service.CommentService, ssoClient *auth.Client) *CommentHandler {
	return &CommentHandler{
		service:   service,
		ssoClient: ssoClient,
	}
}

func (h *CommentHandler) RegisterRoutes(r chi.Router) {
	// Публичные маршруты (не требуют JWT)
	r.Group(func(r chi.Router) {
		r.Get("/comments/{id}", h.getComment)
		r.Get("/products/{productID}/comments", h.getProductComments)
	})

	// Защищенные маршруты (требуют JWT)
	r.Group(func(r chi.Router) {
		r.Use(jwtAuthMiddleware)
		r.Post("/comments", h.createComment)
		r.Put("/comments/{id}", h.updateComment)
		r.Delete("/comments/{id}", h.deleteComment)
	})
}

func jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := jwt.ExtractTokenFromHeader(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *CommentHandler) createComment(w http.ResponseWriter, r *http.Request) {
	var comment models.CreateCommentDTO
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем userID из JWT токена
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	comment.UserID = userID

	id, err := h.service.CreateComment(comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *CommentHandler) getComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	comment, err := h.service.GetComment(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if comment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) getProductComments(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "productID")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	comments, err := h.service.GetProductComments(productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) updateComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var comment models.UpdateCommentDTO
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем userID из JWT токена
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем, принадлежит ли комментарий пользователю
	existingComment, err := h.service.GetComment(id)
	if err != nil {
		http.Error(w, "Failed to get comment", http.StatusInternalServerError)
		return
	}
	if existingComment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Проверяем права доступа
	if existingComment.UserID != userID {
		http.Error(w, "Forbidden: you can only update your own comments", http.StatusForbidden)
		return
	}

	if err := h.service.UpdateComment(id, comment); err != nil {
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CommentHandler) deleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// Получаем userID из JWT токена
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем, принадлежит ли комментарий пользователю
	existingComment, err := h.service.GetComment(id)
	if err != nil {
		http.Error(w, "Failed to get comment", http.StatusInternalServerError)
		return
	}
	if existingComment == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Проверяем права доступа
	if existingComment.UserID != userID {
		http.Error(w, "Forbidden: you can only delete your own comments", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteComment(id); err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
