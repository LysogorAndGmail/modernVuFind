package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/LysogorAndGmail/modernVuFind/internal/repository"
)

func main() {
	// 1. Инициализируем инфраструктурный слой с вашими параметрами Solr
	solrRepo := repository.NewSolrRepository("http://127.0.0.1:8983", "biblio")

	mux := http.NewServeMux()

	// Наш тестовый эндпоинт статуса оставляем для проверки
	mux.HandleFunc("GET /api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "engine": "go-find"})
	})

	// 2. Новый эндпоинт для живого поиска: /api/v1/search?q=ваш_запрос
	mux.HandleFunc("GET /api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			query = "*:*" // Если запрос пустой, отдаем всё (стандарт Solr)
		}

		// Создаем контекст с жестким таймаутом в 2 секунды для защиты сервера от зависаний
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel() // Обязательно освобождаем ресурсы контекста при выходе из функции

		// Вызываем метод поиска из нашего репозитория Solr
		records, err := solrRepo.Search(ctx, query)
		if err != nil {
			log.Printf("Search error for query '%s': %v", query, err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch data from search engine"})
			return
		}

		// Отдаем результат в формате JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(records); err != nil {
			log.Printf("Failed to encode response JSON: %v", err)
		}
	})

	log.Println("API Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
