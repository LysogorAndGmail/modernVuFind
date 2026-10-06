package domain

import "context"

// Record описывает реальный документ из вашего ядра Solr (biblio)
type Record struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	
	// В Solr это массив строк: ["Gundry, Judith M. 1956-"]
	Author    []string `json:"author,omitempty"` 
	
	// В Solr год хранится как строка: "2024"
	Year      string   `json:"year,omitempty"`   
	
	// Тематические фасеты
	Topic     []string `json:"topic_facet,omitempty"`
	
	// Язык книги
	Language  []string `json:"language,omitempty"`
}

// SearchRepository контракт для работы с Solr
type SearchRepository interface {
	Search(ctx context.Context, query string) ([]Record, error)
	BulkIndex(ctx context.Context, records []Record) error
}

