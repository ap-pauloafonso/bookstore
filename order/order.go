package order

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"time"
)

var (
	errEmptyBooksArr        = errors.New("invalid empty books")
	errInvalidBookID        = errors.New("invalid book ID")
	errInvalidBookQuantity  = errors.New("invalid book quantity")
	errDuplicateOrderItemID = errors.New("duplicate bookID, use quantity instead")
)

type Service struct {
	repository  Repository
	bookService BookService
}

func NewService(orderRepository Repository, bookService BookService) *Service {
	return &Service{orderRepository, bookService}
}

type OrderItem struct {
	BookID    int64   `json:"book_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	BookTitle string  `json:"book_title"`
}

type Order struct {
	ID        int64       `json:"id"`
	Total     float64     `json:"total"`
	OrderDate time.Time   `json:"order_date"`
	Items     []OrderItem `json:"items"`
}

func CalculateTotal(items []OrderItem) float64 {
	var r float64

	for _, v := range items {
		r += v.Price * float64(v.Quantity)
	}

	return r
}

type Repository interface {
	SaveOrder(ctx context.Context, customerId int64, orderDate time.Time, items []OrderItem) (*int64, error)
	GetOrdersByCustomer(ctx context.Context, customerID int64) ([]Order, error)
}

type BookService interface {
	GetBooksInformation(ctx context.Context, bookIDs []int64) (map[int64]struct {
		Price float64
		Title string
	}, error)
}

func (s *Service) GetOrdersByCustomer(ctx context.Context, customerID int64) ([]Order, error) {
	orders, err := s.repository.GetOrdersByCustomer(ctx, customerID)
	if err != nil {
		return nil, err
	}

	// calculate total
	distinctBooks := map[int64]struct{}{}
	for i := range orders {
		orders[i].Total = CalculateTotal(orders[i].Items)
		for _, v := range orders[i].Items {
			if _, ok := distinctBooks[v.BookID]; !ok {
				distinctBooks[v.BookID] = struct{}{}
			}
		}
	}

	// fill books name for good user experience
	booksMap, err := s.bookService.GetBooksInformation(ctx, maps.Keys(distinctBooks))
	if err != nil {
		return nil, err
	}
	for i := range orders {
		for bookIdx := range orders[i].Items {
			orders[i].Items[bookIdx].BookTitle = booksMap[orders[i].Items[bookIdx].BookID].Title
		}
	}

	return orders, nil

}

type OrderRequestItem struct {
	BookID   int64 `json:"book_id"`
	Quantity int   `json:"quantity"`
}

func (s *Service) MakeOrder(ctx context.Context, customerID int64, items []OrderRequestItem) (*Order, error) {

	if len(items) == 0 {
		return nil, errEmptyBooksArr
	}

	for _, item := range items {
		if item.BookID <= 0 {
			return nil, errInvalidBookID
		}
		if item.Quantity <= 0 {
			return nil, errInvalidBookQuantity
		}
	}

	// Extract the book IDs from the items
	var bookIDs []int64
	exisitngBookIds := map[int64]struct{}{}
	for _, item := range items {
		if _, ok := exisitngBookIds[item.BookID]; ok {
			return nil, errDuplicateOrderItemID
		}

		exisitngBookIds[item.BookID] = struct{}{}
		bookIDs = append(bookIDs, item.BookID)
	}

	// Check if all books exist, get their prices
	m, err := s.bookService.GetBooksInformation(ctx, bookIDs)
	if err != nil {
		return nil, err
	}

	// fill up the unit prices of each item
	orderItems := make([]OrderItem, len(items))
	for i := range items {
		orderItems[i] = OrderItem{
			BookID:    items[i].BookID,
			Quantity:  items[i].Quantity,
			Price:     m[items[i].BookID].Price,
			BookTitle: m[items[i].BookID].Title,
		}
	}

	t := time.Now()
	// store the order
	orderID, err := s.repository.SaveOrder(ctx, customerID, t, orderItems)
	if err != nil {
		return nil, fmt.Errorf("order creation failed: %w", err)
	}

	return &Order{
		ID:        *orderID,
		Total:     CalculateTotal(orderItems),
		OrderDate: t,
		Items:     orderItems,
	}, nil
}
