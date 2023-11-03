package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/customer"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type CustomerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{db}
}

func (c *CustomerRepository) SaveCustomer(ctx context.Context, email, password string, createdAt time.Time) (*int64, error) {
	var id int64
	err := c.db.QueryRow(ctx, "INSERT INTO customers (email, password, created_at) VALUES ($1, $2, $3) RETURNING id", email, password, createdAt).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("error saving customer: %w", err)
	}

	return &id, nil
}

func (c *CustomerRepository) GetCustomer(ctx context.Context, email string) (*customer.Model, error) {
	var u customer.Model
	err := c.db.QueryRow(ctx, "SELECT id, email, password FROM customers WHERE email = $1", email).Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("error fetching customer: %w", err)
	}

	return &u, nil
}
