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

func TestBookRepository(t *testing.T) {
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

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewBookRepository(pool)

	t.Run("Get books fails because the table doesn't exist yet", func(t *testing.T) {

		_, err := repo.GetAllBooks(context.Background())
		if err == nil {
			t.Fatalf("get all books fails because there no table yet")
		}

	})

	RunMigrations(dsn) // fix db for the next test

	t.Run("Get books works", func(t *testing.T) {

		_, err := repo.GetAllBooks(context.Background())
		if err != nil {
			t.Fatalf("should not return error")
		}

	})
}
