package order

import (
	"context"
	"errors"
	"testing"
	"time"
)

type MockRepository struct {
	SaveOrderFunc           func(ctx context.Context, customerID int64, orderDate time.Time, items []OrderItem) (*int64, error)
	GetOrdersByCustomerFunc func(ctx context.Context, customerID int64) ([]Order, error)
}

func (m *MockRepository) SaveOrder(ctx context.Context, customerID int64, orderDate time.Time, items []OrderItem) (*int64, error) {
	return m.SaveOrderFunc(ctx, customerID, orderDate, items)
}

func (m *MockRepository) GetOrdersByCustomer(ctx context.Context, customerID int64) ([]Order, error) {
	return m.GetOrdersByCustomerFunc(ctx, customerID)
}

type MockBookService struct {
	GetBookPricesFunc func(ctx context.Context, bookIDs []int64) (map[int64]struct {
		Price float64
		Title string
	}, error)
}

func (m *MockBookService) GetBooksInformation(ctx context.Context, bookIDs []int64) (map[int64]struct {
	Price float64
	Title string
}, error) {
	return m.GetBookPricesFunc(ctx, bookIDs)
}

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		name          string
		items         []OrderItem
		expectedTotal float64
	}{
		{
			name:          "SingleItem",
			items:         []OrderItem{{BookID: 1, Quantity: 2, Price: 10.0}},
			expectedTotal: 20.0,
		},
		{
			name:          "MultipleItems",
			items:         []OrderItem{{BookID: 1, Quantity: 2, Price: 10.0}, {BookID: 2, Quantity: 1, Price: 20.0}},
			expectedTotal: 40.0,
		},
		{
			name:          "EmptyItems",
			items:         []OrderItem{},
			expectedTotal: 0.0,
		},
		{
			name:          "MixedItems",
			items:         []OrderItem{{BookID: 1, Quantity: 2, Price: 10.0}, {BookID: 2, Quantity: 1, Price: 20.0}},
			expectedTotal: 40.0,
		},
		{
			name:          "LargeQuantity",
			items:         []OrderItem{{BookID: 1, Quantity: 1000, Price: 0.1}},
			expectedTotal: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := CalculateTotal(tt.items)
			if total != tt.expectedTotal {
				t.Errorf("Expected total: %f, got: %f", tt.expectedTotal, total)
			}
		})
	}
}

func TestService_MakeOrder(t *testing.T) {

	bookServiceErr := errors.New("BookService error")

	saveOrdeErr := errors.New("failed to save")
	tests := []struct {
		name          string
		items         []OrderRequestItem
		expectedOrder *Order
		expectedError error

		SaveOrderFunc           func(ctx context.Context, customerID int64, orderDate time.Time, items []OrderItem) (*int64, error)
		GetOrdersByCustomerFunc func(ctx context.Context, customerID int64) ([]Order, error)

		getBooksInformation func(ctx context.Context, bookIDs []int64) (map[int64]struct {
			Price float64
			Title string
		}, error)
	}{
		{
			name: "ValidOrder",
			items: []OrderRequestItem{
				{BookID: 1, Quantity: 2},
				{BookID: 2, Quantity: 1},
			},
			SaveOrderFunc: func(ctx context.Context, customerID int64, orderDate time.Time, items []OrderItem) (*int64, error) {
				return new(int64), nil
			},
			getBooksInformation: func(ctx context.Context, bookIDs []int64) (map[int64]struct {
				Price float64
				Title string
			}, error) {
				return map[int64]struct {
					Price float64
					Title string
				}{1: {10.0, "Book1"}, 2: {20.0, "Book2"}}, nil
			},
			expectedOrder: &Order{
				ID:        0,
				Total:     40.0,
				OrderDate: time.Time{},
				Items: []OrderItem{
					{BookID: 1, Quantity: 2, Price: 10.0},
					{BookID: 2, Quantity: 1, Price: 20.0},
				},
			},
			expectedError: nil,
		},
		{
			name: "DatabaseFailedToSaveOrder",
			items: []OrderRequestItem{
				{BookID: 1, Quantity: 2},
				{BookID: 2, Quantity: 1},
			},
			SaveOrderFunc: func(ctx context.Context, customerID int64, orderDate time.Time, items []OrderItem) (*int64, error) {
				return new(int64), saveOrdeErr
			},
			getBooksInformation: func(ctx context.Context, bookIDs []int64) (map[int64]struct {
				Price float64
				Title string
			}, error) {
				return map[int64]struct {
					Price float64
					Title string
				}{1: {10.0, "Book1"}, 2: {20.0, "book2"}}, nil
			},
			expectedOrder: &Order{
				ID:        0,
				Total:     40.0,
				OrderDate: time.Time{},
				Items: []OrderItem{
					{BookID: 1, Quantity: 2, Price: 10.0},
					{BookID: 2, Quantity: 1, Price: 20.0},
				},
			},
			expectedError: saveOrdeErr,
		},
		{
			name:  "EmptyItems",
			items: []OrderRequestItem{},

			expectedOrder: nil,
			expectedError: errEmptyBooksArr,
		},
		{
			name: "InvalidBookID",
			items: []OrderRequestItem{
				{BookID: 0, Quantity: 2},
			},
			expectedOrder: nil,
			expectedError: errInvalidBookID,
		},
		{
			name: "InvalidBookQuantity",
			items: []OrderRequestItem{
				{BookID: 1, Quantity: 0},
			},

			expectedOrder: nil,
			expectedError: errInvalidBookQuantity,
		},
		{
			name: "DuplicateBookID",
			items: []OrderRequestItem{
				{BookID: 1, Quantity: 2},
				{BookID: 1, Quantity: 1},
			},

			expectedOrder: nil,
			expectedError: errDuplicateOrderItemID,
		},
		{
			name: "BookServiceError",
			items: []OrderRequestItem{
				{BookID: 1, Quantity: 2},
				{BookID: 2, Quantity: 1},
			},
			getBooksInformation: func(ctx context.Context, bookIDs []int64) (map[int64]struct {
				Price float64
				Title string
			}, error) {
				return nil, bookServiceErr
			},
			expectedOrder: nil,
			expectedError: bookServiceErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create a mock repository and book service.
			mockRepo := &MockRepository{
				SaveOrderFunc:           tt.SaveOrderFunc,
				GetOrdersByCustomerFunc: tt.GetOrdersByCustomerFunc,
			}
			mockBookService := &MockBookService{
				GetBookPricesFunc: tt.getBooksInformation,
			}

			// Create the service with the mock repository and book service.
			service := NewService(mockRepo, mockBookService)

			order, err := service.MakeOrder(context.Background(), 1, tt.items)

			if err != nil {
				if tt.expectedError == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
			} else if tt.expectedError != nil {
				t.Errorf("Expected error: %v, got nil", tt.expectedError)
			}

			if order != nil {
				if tt.expectedOrder == nil {
					t.Errorf("Expected nil order, got: %v", order)
				} else {
					if order.ID != tt.expectedOrder.ID {
						t.Errorf("Expected ID: %v, got: %v", tt.expectedOrder.ID, order.ID)
					}
					if order.Total != tt.expectedOrder.Total {
						t.Errorf("Expected total: %f, got: %f", tt.expectedOrder.Total, order.Total)
					}
				}
			}
		})
	}
}

func TestService_GetOrdersByCustomer(t *testing.T) {
	repoErr := errors.New("repo err")
	booksInfoErr := errors.New("error retrieving books info")
	// Prepare some sample data for testing
	orders := []Order{
		{
			ID:        1,
			OrderDate: time.Now(),
			Items: []OrderItem{
				{BookID: 1, Quantity: 2, Price: 10.0},
				{BookID: 2, Quantity: 1, Price: 20.0},
			},
		},
		{
			ID:        2,
			OrderDate: time.Now(),
			Items: []OrderItem{
				{BookID: 3, Quantity: 3, Price: 30.0},
				{BookID: 4, Quantity: 2, Price: 40.0},
			},
		},
	}

	tests := []struct {
		name                string
		customerID          int64
		mockRepositoryFunc  func(ctx context.Context, customerID int64) ([]Order, error)
		expectedOrders      []Order
		expectedError       error
		getBooksInformation func(ctx context.Context, bookIDs []int64) (map[int64]struct {
			Price float64
			Title string
		}, error)
	}{
		{
			name:       "ValidCustomer",
			customerID: 1,
			mockRepositoryFunc: func(ctx context.Context, customerID int64) ([]Order, error) {
				return orders, nil
			},
			expectedOrders: orders, // Orders should be the same as provided.
			expectedError:  nil,
			getBooksInformation: func(ctx context.Context, bookIDs []int64) (map[int64]struct {
				Price float64
				Title string
			}, error) {
				return map[int64]struct {
					Price float64
					Title string
				}{1: {10.0, "book1"}, 2: {20.0, "book2"}}, nil
			},
		},
		{
			name:       "InvalidCustomer",
			customerID: 999,
			mockRepositoryFunc: func(ctx context.Context, customerID int64) ([]Order, error) {
				return nil, repoErr
			},
			expectedOrders: nil,
			expectedError:  repoErr,
		},
		{
			name:       "failed to fill books name",
			customerID: 1,
			mockRepositoryFunc: func(ctx context.Context, customerID int64) ([]Order, error) {
				return orders, nil
			},
			expectedError: booksInfoErr,
			getBooksInformation: func(ctx context.Context, bookIDs []int64) (map[int64]struct {
				Price float64
				Title string
			}, error) {
				return nil, booksInfoErr
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock repository.
			mockRepo := &MockRepository{
				GetOrdersByCustomerFunc: tt.mockRepositoryFunc,
			}

			mockBookService := &MockBookService{
				GetBookPricesFunc: tt.getBooksInformation,
			}

			// Create the service with the mock repository.
			service := NewService(mockRepo, mockBookService)

			orders, err := service.GetOrdersByCustomer(context.Background(), tt.customerID)

			if err != nil {
				if tt.expectedError == nil || !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
			} else if tt.expectedError != nil {
				t.Errorf("Expected error: %v, got nil", tt.expectedError)
			}

			if len(orders) != len(tt.expectedOrders) {
				t.Errorf("Expected %d orders, got %d", len(tt.expectedOrders), len(orders))
			} else {
				// Check individual orders if needed.
				for i := range orders {
					if orders[i].ID != tt.expectedOrders[i].ID {
						t.Errorf("Order %d: Expected ID %d, got %d", i, tt.expectedOrders[i].ID, orders[i].ID)
					}
					if orders[i].Total != tt.expectedOrders[i].Total {
						t.Errorf("Order %d: Expected Total %f, got %f", i, tt.expectedOrders[i].Total, orders[i].Total)
					}
					// Check other fields and items as needed.
				}
			}
		})
	}
}
