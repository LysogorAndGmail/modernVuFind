package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/LysogorAndGmail/modernVuFind/internal/domain"
	"net/http"
	"net/url"
)

// SolrRepository имплементирует интерфейс domain.SearchRepository
type SolrRepository struct {
	client  *http.Client
	baseURL string
	core    string
}

// NewSolrRepository — конструктор репозитория
func NewSolrRepository(baseURL, core string) *SolrRepository {
	return &SolrRepository{
		client:  &http.Client{}, // Под капотом автоматический пул коннектов
		baseURL: baseURL,
		core:    core,
	}
}

// Search выполняет поисковый запрос к Solr и парсит результат
// Search выполняет поисковый запрос к Solr и парсит результат
func (r *SolrRepository) Search(ctx context.Context, query string) ([]domain.Record, error) {
	params := url.Values{}
	params.Add("q", query)
	params.Add("wt", "json")
        params.Add("fl", "*")
        params.Add("rows", "20")

	solrURL := fmt.Sprintf("%s/solr/%s/select?%s", r.baseURL, r.core, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, solrURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("solr request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("solr returned status %d", resp.StatusCode)
	}

	// ─── ДОБАВЛЯЕМ ОТЛАДКУ СЮДА ───
	// Читаем тело ответа в буфер, чтобы прочитать его дважды (для лога и для парсинга)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// БЕЗОПАСНЫЙ ЛОГ: проверяем длину, чтобы не было паники slice bounds out of range
	responseStr := buf.String()
	limit := 500
	if len(responseStr) < limit {
		limit = len(responseStr)
	}
	log.Printf("[DEBUG SOLR RESPONSE]: %s", responseStr[:limit])
	// ──────────────────────────────

	var solrResponse struct {
		Response struct {
			Docs []domain.Record `json:"docs"`
		} `json:"response"`
	}

	// Декодируем из буфера
	if err := json.NewDecoder(&buf).Decode(&solrResponse); err != nil {
		return nil, fmt.Errorf("failed to decode solr response: %w", err)
	}

	return solrResponse.Response.Docs, nil
}




// BulkIndex отправляет пачку документов в Solr на индексацию
func (r *SolrRepository) BulkIndex(ctx context.Context, records []domain.Record) error {
	// commitWithin=1000 говорит Solr применить изменения в течение 1 секунды
	solrURL := fmt.Sprintf("%s/solr/%s/update?commitWithin=1000", r.baseURL, r.core)

	// Сериализуем пачку записей в JSON
	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, solrURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create post request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("solr bulk index request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("solr index returned status %d", resp.StatusCode)
	}

	return nil
}
