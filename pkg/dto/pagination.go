package dto

import (
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Cursor  string
	Limit   int
}

const (
	_defaultLimit = 20
	_maxLimit     = 100
)

func ParsePagination(r *http.Request) PaginationParams {
	p := PaginationParams{
		Cursor: r.URL.Query().Get("cursor"),
		Limit:  _defaultLimit,
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			p.Limit = parsed
		}
	}

	if p.Limit > _maxLimit {
		p.Limit = _maxLimit
	}

	return p
}
