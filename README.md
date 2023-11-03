# bookstore

## Assumptions 
1. `A customer needs to see my orders history`: it's unclear whether the customer should see all orders or just his own orders, so I'm assuming that each customer can only see his own orders
2. To see a list of available books, I'm not requiring authentication, but if the user wants to make an order he must authenticate first


## Running 
1. `make docker-run`
2. docs at http://localhost:8081/swagger/

## Authentication
* Use `Authorization` header with `Bearer <TOKEN>`

## Endpoints
* `GET /health` api health endpoint
* `POST /api/register` api for registering new customer (returns an JWT TOKEN)
* `POST /api/login` api for customer login (returns an JWT TOKEN)
* `GET /api/books` api for listing the available books (doesn't require authentication)
* `POST /api/orders` api for creating an order (requires authentication)
* `GET /api/orders` api for listing customer orders (requires authentication)


## Tests
* `make test-unit` for unit tests (business layer) - 100% coverage
* `make test-all` for unit tests + integration tests (storage layer) - 87% coverage 
* `make test-e2e` for e2e test

## Test Coverage Map
* `make coverage-unit` to open coverage map from unit tests
* `make coverage-all`  to open coverage map from unit tests from unit tests + integration tests
