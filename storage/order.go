package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/order"
	"github.com/jackc/pgx/v4/pgxpool"
	"sort"
	"time"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db}
}

func (r *OrderRepository) SaveOrder(ctx context.Context, customerID int64, orderDate time.Time, items []order.OrderItem) (*int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert an order record
	var orderID int64 // Change the data type to int64
	orderInsertSQL := "INSERT INTO orders (customer_id, create_id) VALUES ($1, $2) RETURNING id"
	if err := tx.QueryRow(ctx, orderInsertSQL, customerID, orderDate).Scan(&orderID); err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	// Insert order items
	orderItemInsertSQL := "INSERT INTO orderitems (order_id, book_id, quantity, price) VALUES ($1, $2, $3, $4)"
	for _, item := range items {
		if _, err := tx.Exec(ctx, orderItemInsertSQL, orderID, item.BookID, item.Quantity, item.Price); err != nil {
			return nil, fmt.Errorf("error adding order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &orderID, nil // Return the order ID as int64
}

func (r *OrderRepository) GetOrdersByCustomer(ctx context.Context, customerID int64) ([]order.Order, error) {
	query := `
        SELECT o.id, o.create_id, oi.book_id, b.title, oi.quantity, oi.price
        FROM orders o
        JOIN orderitems oi ON o.id = oi.order_id
		JOIN books b ON b.id = oi.book_id
        WHERE o.customer_id = $1
    `

	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a map to group orders by their IDs
	orderMap := make(map[int64]*order.Order)

	for rows.Next() {

		var orderItem order.OrderItem
		var o order.Order
		if err := rows.Scan(&o.ID, &o.OrderDate, &orderItem.BookID, &orderItem.BookTitle, &orderItem.Quantity, &orderItem.Price); err != nil {
			return nil, err
		}

		// Check if the order exists in the map, if not, create a new one
		if existingOrder, ok := orderMap[o.ID]; ok {
			existingOrder.Items = append(existingOrder.Items, orderItem)
		} else {
			newOrder := &order.Order{
				ID:        o.ID,
				OrderDate: o.OrderDate,
				Items:     []order.OrderItem{orderItem},
			}
			orderMap[o.ID] = newOrder
		}
	}

	// Convert the map values (orders) into a slice
	orders := []order.Order{}
	for _, o := range orderMap {
		orders = append(orders, *o)
	}

	// Sort the orders in descending order by Order ID
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].ID > orders[j].ID
	})

	return orders, nil
}
