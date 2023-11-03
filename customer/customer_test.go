package customer

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Define a mock repository for testing purposes.
type MockRepository struct {
	customers map[string]*Model
	Err       error
}

func (m *MockRepository) SaveCustomer(ctx context.Context, email, password string, createdAt time.Time) (*int64, error) {
	if _, exists := m.customers[email]; exists {
		return nil, errors.New("customer already exists")
	}
	m.customers[email] = &Model{

		Email:    email,
		Password: password,
	}
	var n int64
	return &n, m.Err
}

func (m *MockRepository) GetCustomer(ctx context.Context, email string) (*Model, error) {
	customer, exists := m.customers[email]
	if !exists {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

// Define a mock repository for testing purposes.
type MockSecurity struct {
	errorHash   error
	resultCheck bool
}

func (m *MockSecurity) HashPassword(password string) (string, error) {
	return "", m.errorHash
}

func (m *MockSecurity) CheckPasswordHash(password, hash string) bool {

	return m.resultCheck
}

func TestService_Register(t *testing.T) {
	errHash := errors.New("error hashing")
	// Define test cases as a table.
	testCases := []struct {
		name         string
		email        string
		password     string
		repoError    error
		expectedErr  error
		errorHashing error
		customers    map[string]*Model
	}{
		{
			name:        "Valid registration",
			email:       "user@example.com",
			password:    "password",
			expectedErr: nil,
			customers:   map[string]*Model{},
		},
		{
			name:        "Short password",
			email:       "user@example.com",
			password:    "pw",
			expectedErr: errPasswordShort,
			customers:   map[string]*Model{},
		},
		{
			name:        "Long password",
			email:       "user@example.com",
			password:    "ThisIsAVeryLongPasswordThatExceedsTheMaximumLengthOfFiftyCharacters",
			expectedErr: errPasswordLong,
			customers:   map[string]*Model{},
		},
		{
			name:        "Long email",
			email:       "ThisIsAReallyLongEmailAddressThatExceedsTheMaximumLengthOfTwoHundredAndFiftyFiveCharactersssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss@example.com",
			password:    "password",
			expectedErr: errEmailLong,
			customers:   map[string]*Model{},
		},
		{
			name:        "Invalid email",
			email:       "invalid_email",
			password:    "password",
			expectedErr: errEmailInvalid,
			customers:   map[string]*Model{},
		},
		{
			name:     "Email already taken",
			email:    "user@example.com",
			password: "password",
			customers: map[string]*Model{
				"user@example.com": {
					Id:       1,
					Email:    "user@example.com",
					Password: "password",
				},
			},
			repoError:   errEmailAlreadyTaken,
			expectedErr: errEmailAlreadyTaken,
		},
		{
			name:         "Error hashing password",
			email:        "user@example.com",
			password:     "password",
			errorHashing: errHash,
			expectedErr:  errHash,
		},
		{
			name:        "Error storing customer",
			email:       "user@example.com",
			password:    "password",
			repoError:   errStoringcustomer,
			expectedErr: errStoringcustomer,
			customers:   map[string]*Model{},
		},
	}

	// Iterate through the test cases and run the tests.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			service := NewService(&MockRepository{
				customers: tc.customers,
				Err:       tc.repoError,
			}, &MockSecurity{
				errorHash: tc.errorHashing,
			})

			_, err := service.Register(context.Background(), tc.email, tc.password)

			// Check the error.
			if err != tc.expectedErr {
				t.Fatalf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestLogin(t *testing.T) {

	service := NewService(&MockRepository{
		customers: map[string]*Model{},
		Err:       nil,
	}, &MockSecurity{
		errorHash:   nil,
		resultCheck: true,
	})
	_, err := service.Register(context.Background(), "user@gmail.com", "password5")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Valid Login", func(t *testing.T) {
		_, err := service.Login(context.Background(), "user@gmail.com", "password5")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	service2 := NewService(&MockRepository{
		customers: map[string]*Model{
			"use2r@gmail.com": {
				Id:       1,
				Email:    "user@gmail.com",
				Password: "pass",
			},
		},
		Err: nil,
	}, &MockSecurity{
		errorHash:   nil,
		resultCheck: true,
	})
	t.Run("Failed get", func(t *testing.T) {
		_, err := service2.Login(context.Background(), "user@gmail.com", "pass")
		if err != errInvalidCredentials {
			t.Errorf("Expected %v, got %v", errInvalidCredentials, err)
		}
	})

	service3 := NewService(&MockRepository{
		customers: map[string]*Model{
			"user@gmail.com": {
				Id:       1,
				Email:    "user@gmail.com",
				Password: "pass",
			},
		},
		Err: nil,
	}, &MockSecurity{
		errorHash:   nil,
		resultCheck: false,
	})
	t.Run("Invalid Credentials", func(t *testing.T) {
		_, err := service3.Login(context.Background(), "user@gmail.com", "pass")
		if err != errInvalidCredentials {
			t.Errorf("Expected %v, got %v", errInvalidCredentials, err)
		}
	})

}

func TestService_Getcustomer(t *testing.T) {

	t.Run("get customer works", func(t *testing.T) {
		service := NewService(&MockRepository{
			customers: map[string]*Model{
				"user@gmail.com": {
					Id:       1,
					Email:    "user@gmail.com",
					Password: "pass",
				},
			},
			Err: nil,
		}, nil)

		_, err := service.Getcustomer(context.Background(), "user@gmail.com")
		if err != nil {
			t.Fatalf("expected no error and got %v", err)
		}
	})

	t.Run("get customer fails", func(t *testing.T) {
		service := NewService(&MockRepository{
			customers: map[string]*Model{
				"user@gmail.com": {
					Id:       1,
					Email:    "user@gmail.com",
					Password: "pass",
				},
			},
			Err: nil,
		}, nil)

		_, err := service.Getcustomer(context.Background(), "user2@gmail.com")
		if err == nil {
			t.Fatalf("expected error and got nil")
		}
	})

}
