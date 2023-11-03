package book

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// Define a mock repository for testing purposes.
type MockRepository struct {
	Books []*Model
	Err   error
}

func (m *MockRepository) GetAllBooks(ctx context.Context) ([]*Model, error) {
	return m.Books, m.Err
}

func TestService_GetAllBooks(t *testing.T) {
	errRepo := errors.New("mock repository error")
	testCases := []struct {
		name          string
		mockBooks     []*Model
		expectedBooks []*Model
		mockError     error
		expectedError error
	}{
		{
			name: "Valid list of books",
			mockBooks: []*Model{
				{
					ID:     1,
					Title:  "harry potter",
					Author: "jk rowling",
					Price:  10,
				},
			},
			expectedBooks: []*Model{
				{
					ID:     1,
					Title:  "harry potter",
					Author: "jk rowling",
					Price:  10,
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "Error from repository",
			mockBooks:     nil,
			mockError:     errRepo,
			expectedError: errRepo,
		},
	}

	// Iterate through the test cases and run the tests.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a Service with a mock repository.
			service := NewService(&MockRepository{
				Books: tc.mockBooks,
				Err:   tc.mockError,
			})

			// Perform the actual test.
			books, err := service.GetAllBooks(context.Background())

			if !reflect.DeepEqual(books, tc.expectedBooks) {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedBooks, books)

			}
			// Check the error.
			if err != tc.expectedError {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedError, err)
			}
		})
	}
}

func TestService_GetBookPrices(t *testing.T) {
	errRepo := errors.New("mock repository error")
	testCases := []struct {
		name          string
		bookIDs       []int64
		bookOptions   []*Model
		expected      map[int64]float64
		repoError     error
		expectedError error
	}{
		{
			name:    "All book IDs exist",
			bookIDs: []int64{1, 2, 3},
			bookOptions: []*Model{
				{
					ID:     1,
					Title:  "book1",
					Author: "author1",
					Price:  1.0,
				},
				{
					ID:     2,
					Title:  "book2",
					Author: "author2",
					Price:  2.0,
				},
				{
					ID:     3,
					Title:  "book3",
					Author: "author3",
					Price:  3.0,
				},
				{
					ID:     4,
					Title:  "book4",
					Author: "author4",
					Price:  4.0,
				},
			},
			expected:      map[int64]float64{1: 1.0, 2: 2.0, 3: 3.0},
			repoError:     nil,
			expectedError: nil,
		},
		{
			name:    "One of the book IDs is not found",
			bookIDs: []int64{1, 4, 3},
			bookOptions: []*Model{
				{
					ID:     1,
					Title:  "book1",
					Author: "author1",
					Price:  1.0,
				},
				{
					ID:     2,
					Title:  "book2",
					Author: "author2",
					Price:  2.0,
				},
				{
					ID:     3,
					Title:  "book3",
					Author: "author3",
					Price:  3.0,
				},
			},
			expected:      nil,
			repoError:     nil,
			expectedError: errBookNotFound,
		},
		{
			name:          "Error from repository",
			bookIDs:       []int64{1, 2, 3},
			expected:      nil,
			repoError:     errRepo,
			expectedError: errRepo,
		},
	}

	// Iterate through the test cases and run the tests.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Create a Service with a mock repository.
			service := NewService(&MockRepository{
				Books: tc.bookOptions,
				Err:   tc.repoError,
			})

			prices, err := service.GetBookPrices(context.Background(), tc.bookIDs)

			// Check the error.
			if err != tc.expectedError {
				t.Fatalf("Expected error: %v, but got: %v", tc.repoError, err)
			}

			// Check the prices if there's no error.
			if err == nil {
				for id, expectedPrice := range tc.expected {
					if prices[id] != expectedPrice {
						t.Errorf("Expected price for book %d: %f, but got: %f", id, expectedPrice, prices[id])
					}
				}
			}
		})
	}
}
