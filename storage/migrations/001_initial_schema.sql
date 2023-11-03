-- +goose Up
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id),
    create_id TIMESTAMP NOT NULL
);

CREATE TABLE orderitems (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id),
    book_id INT REFERENCES books(id),
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS orderitems;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS users;