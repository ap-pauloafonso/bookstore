package book

import (
	"context"
	"errors"
)

var (
	errBookNotFound = errors.New("book not found")
)

type Service struct {
	r Repository
}

func NewService(bookRepository Repository) *Service {
	return &Service{bookRepository}
}

type Model struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

type Repository interface {
	GetAllBooks(ctx context.Context) ([]*Model, error)
}

func (s *Service) GetAllBooks(ctx context.Context) ([]*Model, error) {
	return s.r.GetAllBooks(ctx)
}

// GetBookPrices returns a map of book prices if all of them exists, and errBookNotFound if one of the books is not found
func (s *Service) GetBookPrices(ctx context.Context, bookIDs []int64) (map[int64]float64, error) {

	books, err := s.GetAllBooks(ctx)
	if err != nil {
		return nil, err
	}

	m := map[int64]float64{}

	for _, v := range books {
		m[v.ID] = v.Price
	}

	for _, v := range bookIDs {
		if _, ok := m[v]; !ok {
			return nil, errBookNotFound
		}
	}

	return m, nil
}
