package storage

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/order"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

func TestOrderRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Parallel()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "MY_DB",
		},
		WaitingFor: wait.ForAll(wait.ForListeningPort("5432/tcp"), wait.ForLog("database system is ready to accept connections")),
	}
	postgresC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	defer postgresC.Terminate(context.Background())

	host, err := postgresC.Host(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	port, err := postgresC.MappedPort(context.Background(), "5432")
	if err != nil {
		t.Error(err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=test dbname=MY_DB sslmode=disable", host, port.Port())

	time.Sleep(3 * time.Second) // a bit of delay to make sure that container is ready

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewOrderRepository(pool)

	customerRepo := NewCustomerRepository(pool)

	t.Run("get orders fails because there is no table yet", func(t *testing.T) {

		_, err := repo.GetOrdersByCustomer(context.Background(), 1)
		if err == nil {
			t.Fatalf("shoould have an error because there is no table created yet")
		}

	})

	t.Run("save order fails because the is no table yet", func(t *testing.T) {

		_, err := repo.SaveOrder(context.Background(), 1, time.Now(), []order.OrderItem{
			{BookID: 1, Quantity: 1, Price: 5},
			{BookID: 2, Quantity: 10, Price: 7},
			{BookID: 3, Quantity: 30, Price: 9},
		})
		if err == nil {
			t.Fatalf("shoould have an error because there is no table created yet")
		}

	})

	err = RunMigrations(dsn)
	if err != nil {
		t.Fatal(err)
	}

	var customerid *int64

	t.Run(" save order works", func(t *testing.T) {

		customerid, err = customerRepo.SaveCustomer(context.Background(), "test@gmail.com", "123", time.Now())
		if err != nil {
			t.Fatalf("should not have an error while creating a customer to create order later")
		}

		_, err := repo.SaveOrder(context.Background(), *customerid, time.Now(), []order.OrderItem{
			{BookID: 1, Quantity: 1, Price: 5},
			{BookID: 2, Quantity: 10, Price: 7},
			{BookID: 3, Quantity: 30, Price: 9},
		})
		if err != nil {
			t.Fatalf("should not have an error while inserting the order")
		}

	})

	t.Run("get orders works", func(t *testing.T) {

		o, err := repo.GetOrdersByCustomer(context.Background(), *customerid)
		if err != nil {
			t.Fatalf("should not have an error while hetting the orders")
		}

		if len(o) != 1 {
			t.Fatalf("should have 1 order in the result")
		}

		if len(o[0].Items) != 3 {
			t.Fatalf("should have 3 items in the order")

		}
	})

}
