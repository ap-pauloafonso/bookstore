package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

func TestCustomerRepository(t *testing.T) {
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

	time.Sleep(3 * time.Second) // a bit of delay to make sure that container is ready
	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=test dbname=MY_DB sslmode=disable", host, port.Port())

	err = RunMigrations(dsn)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewCustomerRepository(pool)

	t.Run("Savecustomer", func(t *testing.T) {

		_, err := repo.SaveCustomer(context.Background(), "test@gmail.com", "123456", time.Now())
		if err != nil {
			t.Fatalf("should not have error while saving new customer")
		}

	})

	t.Run("Savecustomer fails because the password is too long", func(t *testing.T) {

		_, err := repo.SaveCustomer(context.Background(), "error@gmail.com", "longpassword_longpassword_longpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpasswordlongpassword_longpassword", time.Now())
		if err == nil {
			t.Fatalf("should have an error saving customer")
		}

	})

	t.Run("Getcustomer", func(t *testing.T) {
		customerSave, err := repo.SaveCustomer(context.Background(), "test2@gmail.com", "123456", time.Now())
		if err != nil {
			t.Fatalf("should not have error while saving new customer")
		}

		customerGET, err := repo.GetCustomer(context.Background(), "test2@gmail.com")
		if err != nil {
			t.Fatalf("should not have error while querying the customer")
		}

		if *customerSave != customerGET.Id {
			t.Fatalf("customer id from query should be the same the inserted id")
		}
	})

	t.Run("Getcustomer fails because no customer is found", func(t *testing.T) {
		_, err := repo.GetCustomer(context.Background(), "notfound@gmail.com")
		if err == nil {
			t.Fatalf("should have an error while retriving customer")
		}

	})

}
