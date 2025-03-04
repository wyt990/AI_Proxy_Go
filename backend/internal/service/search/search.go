package search

import (
	m "AI_Proxy_Go/backend/internal/model"
	"time"
)

type SearchService interface {
	Search(query string) (*m.SearchResult, error)
	ProcessQuery(query string) (string, error)
	FilterResults(results *m.SearchResult) (*m.SearchResult, error)
	CheckRateLimit() (bool, time.Duration)
}
