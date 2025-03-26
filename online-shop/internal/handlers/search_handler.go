package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"online-shop/internal/services"

	"github.com/go-chi/chi/v5"
)

type SearchHandler struct {
	Service *services.SearchService
}

func NewSearchHandler(service *services.SearchService) *SearchHandler {
	return &SearchHandler{Service: service}
}

func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	products, err := h.Service.SearchProducts(query)
	if err != nil {
		log.Printf("Ошибка при поиске товаров, я вылетаю из серч хендлера: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *SearchHandler) SetupRoutes(r chi.Router) {
	r.Get("/search", h.HandleSearch)
}
