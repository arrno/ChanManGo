package domain

import (
	"errors"
	"net/http"
	"strconv"
)

type Pagination[T any] struct {
	Page  int `json:"page"`
	Count int `json:"count"`
	Found int `json:"found"`
	Total int `json:"total"`
	Items []T `json:"items"`
}

func (p *Pagination[T]) Validate() error {
	if p.Page < 1 {
		return errors.New("Invalid pagination. Page should be a positive non zero number.")
	}
	if p.Count < 1 {
		return errors.New("Invalid pagination. Count should be a positive non zero number.")
	}
	return nil
}

func (p *Pagination[T]) DoPaginate(values []T) []T {
	start := (p.Page - 1) * p.Count
	if start >= len(values) {
		return make([]T, 0)
	}
	stop := min(len(values), start+p.Count)
	return values[start:stop]
}

func (p *Pagination[T]) FromReq(r *http.Request) error {
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		return err
	}
	count, err := strconv.ParseInt(r.URL.Query().Get("count"), 10, 64)
	if err != nil {
		return err
	}
	p.Page = int(page)
	p.Count = int(count)
	return nil
}
