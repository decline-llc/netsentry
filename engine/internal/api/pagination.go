package api

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

type pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func parsePagination(r *http.Request) (pagination, error) {
	page, err := parsePositiveInt(r, "page", defaultPage)
	if err != nil {
		return pagination{}, err
	}
	perPage, err := parsePositiveInt(r, "per_page", defaultPerPage)
	if err != nil {
		return pagination{}, err
	}
	if perPage > maxPerPage {
		return pagination{}, fmt.Errorf("per_page must be <= %d", maxPerPage)
	}
	// Validate before either backend computes (page-1)*perPage.
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/perPage {
		return pagination{}, fmt.Errorf("page and per_page exceed maximum pagination offset")
	}
	return pagination{Page: page, PerPage: perPage}, nil
}

func parsePositiveInt(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func pageBounds(total int, p pagination) (int, int) {
	start := (p.Page - 1) * p.PerPage
	if start >= total {
		return total, total
	}
	// Clamp the length before adding so even a MaxInt total cannot overflow.
	length := total - start
	if length > p.PerPage {
		length = p.PerPage
	}
	return start, start + length
}
