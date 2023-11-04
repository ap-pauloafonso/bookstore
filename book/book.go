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

// GetBooksInformation returns a map of book price/name if all of them exists, and errBookNotFound if one of the books is not found
func (s *Service) GetBooksInformation(ctx context.Context, bookIDs []int64) (map[int64]struct {
	Price float64
	Title string
}, error) {

	books, err := s.GetAllBooks(ctx)
	if err != nil {
		return nil, err
	}

	m := map[int64]struct {
		Price float64
		Title string
	}{}

	for _, v := range books {
		m[v.ID] = struct {
			Price float64
			Title string
		}{Price: v.Price, Title: v.Title}
	}

	for _, v := range bookIDs {
		if _, ok := m[v]; !ok {
			return nil, errBookNotFound
		}
	}

	return m, nil
}
