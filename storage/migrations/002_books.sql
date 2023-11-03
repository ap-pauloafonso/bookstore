-- +goose Up
INSERT INTO books (title, author, price)
VALUES
    ('Harry Potter and the Sorcerer''s Stone', 'J.K. Rowling', 10.99),
    ('Harry Potter and the Chamber of Secrets', 'J.K. Rowling', 11.99),
    ('Harry Potter and the Prisoner of Azkaban', 'J.K. Rowling', 12.99),
    ('Harry Potter and the Goblet of Fire', 'J.K. Rowling', 13.99),
    ('Harry Potter and the Order of the Phoenix', 'J.K. Rowling', 14.99),
    ('Harry Potter and the Half-Blood Prince', 'J.K. Rowling', 15.99),
    ('Harry Potter and the Deathly Hallows', 'J.K. Rowling', 16.99),
    ('Fantastic Beasts and Where to Find Them', 'J.K. Rowling', 9.99),
    ('Quidditch Through the Ages', 'J.K. Rowling', 8.99),
    ('The Tales of Beedle the Bard', 'J.K. Rowling', 7.99);


-- +goose Down
DELETE FROM books WHERE author = 'J.K. Rowling'
