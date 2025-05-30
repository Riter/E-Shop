package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Riter/E-Shop/internal/models"
)

type Source interface{
	GetProductsByIDs(ctx context.Context, skus []int64) ([]models.ProductResponse, error)
}

type Cacher interface{
	Get(ctx context.Context, keys ...string) ([]interface{}, error)
	Set(ctx context.Context, mset map[string]string, expiration time.Duration)
}

func GetProductsWithCache(ctx context.Context, skus []int64, source Source, cacher Cacher) (models.ProductResponseList, error) {
	keys := make([]string, len(skus))
	for i, id := range skus {
		keys[i] = strconv.FormatInt(id, 10)
	}

	// 1. Batch GET из Redis
	cached, err := cacher.Get(ctx, keys...)
	if err != nil {
		return models.ProductResponseList{}, err
	}

	var found []models.ProductResponse
	var missedIDs []int64
	mset := make(map[string]string) // для последующего кэширования

	for i, val := range cached {
		if val == nil {
			missedIDs = append(missedIDs, skus[i])
			continue
		}

		var p models.ProductResponse
		if err := json.Unmarshal([]byte(val.(string)), &p); err != nil {
			// Не добавляем если ошибка
			return models.ProductResponseList{}, err
		}
		found = append(found, p)
	}

	// 2. Если есть пропущенные — добираем из БД
	if len(missedIDs) > 0 {
		dbResults, err := source.GetProductsByIDs(ctx, missedIDs)
		if err != nil {
			return models.ProductResponseList{}, err
		}

		for _, p := range dbResults {
			found = append(found, p)

			// Сериализуем и готовим к Set
			raw, err := json.Marshal(p)
			if err!=nil{
				return models.ProductResponseList{}, err
			}
			key := strconv.FormatInt(int64(p.ID), 10)
			mset[key] = string(raw)
		}

		// 3. Batch SET в Redis с TTL (пример: 5 минут)
		ctx = context.WithoutCancel(ctx)
		log.Print("mset:", mset)
		go cacher.Set(ctx, mset, time.Minute*5)

	}

	return models.ProductResponseList{ProductList: found}, nil
}


func GetProducts(ctx context.Context, source Source, cacher Cacher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.ProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.SKUs) == 0 {
			http.Error(w, "invalid or empty request", http.StatusBadRequest)
			return
		}

		result, err := GetProductsWithCache(r.Context(), req.SKUs, source, cacher)
		if err != nil {
			http.Error(w, "server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
