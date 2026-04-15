package helpers

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 200
)

// Pagination содержит параметры постраничной навигации.
type Pagination struct {
	Page  int // Номер страницы, начиная с 1
	Limit int // Размер страницы
}

// Offset возвращает смещение для SQL-запроса.
func (p Pagination) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// ParsePagination читает ?page=N&limit=M из запроса.
// Применяет разумные значения по умолчанию и ограничения.
func ParsePagination(r *http.Request) Pagination {
	page := 1
	limit := DefaultPageSize

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > MaxPageSize {
				n = MaxPageSize
			}
			limit = n
		}
	}
	return Pagination{Page: page, Limit: limit}
}

// PagedResponse — обёртка для постраничного ответа API.
type PagedResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	HasMore bool `json:"hasMore"`
}

// NewPagedResponse создаёт постраничный ответ.
func NewPagedResponse[T any](items []T, total int, p Pagination) PagedResponse[T] {
	if items == nil {
		items = []T{}
	}
	return PagedResponse[T]{
		Items:   items,
		Total:   total,
		Page:    p.Page,
		Limit:   p.Limit,
		HasMore: p.Offset()+len(items) < total,
	}
}
