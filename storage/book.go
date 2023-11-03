package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/book"
	"github.com/jackc/pgx/v4/pgxpool"
)

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{db}
}

func (r *BookRepository) GetAllBooks(ctx context.Context) ([]*book.Model, error) {
	rows, err := r.db.Query(ctx, "SELECT id, title, author, price FROM books")
	if err != nil {
		return nil, fmt.Errorf("error fetching books: %w", err)
	}
	defer rows.Close()

	var books []*book.Model
	for rows.Next() {
		var b book.Model
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price)
		if err != nil {
			return nil, err
		}
		books = append(books, &b)
	}

	return books, nil
}
