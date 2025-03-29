package domain

import "errors"

type Pagination[T any] struct {
	Page  int `json:"page"`
	Count int `json:"count"`
	Found int `json:"found"`
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
