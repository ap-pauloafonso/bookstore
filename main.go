package main

import (
	"context"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/book"
	"github.com/ap-pauloafonso/bookstore/config"
	"github.com/ap-pauloafonso/bookstore/customer"
	"github.com/ap-pauloafonso/bookstore/order"
	"github.com/ap-pauloafonso/bookstore/security"
	"github.com/ap-pauloafonso/bookstore/server"
	"github.com/ap-pauloafonso/bookstore/storage"
	"github.com/ap-pauloafonso/bookstore/utils"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lmittmann/tint"
	"github.com/sethvargo/go-envconfig"

	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.Info("starting the server...")

	ctx := context.Background()

	// try to load env vars from .env file (useful when running locally)
	godotenv.Load()

	// use tint to give some color to the slogs output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	// process/validate env variables
	var cfg config.GlobalConfig
	if err := envconfig.Process(ctx, &cfg); err != nil {
		utils.LogErrorFatal(fmt.Errorf("missing env vars - %w", err))
	}

	// setup channel for listening to events from OS
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	//perform migration
	err := storage.RunMigrations(cfg.PostgresConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Initialize the database connection pool
	db, err := pgxpool.Connect(ctx, cfg.PostgresConnection)
	if err != nil {
		utils.LogErrorFatal(err)
	}

	// Close the database connection pool when the application exits
	defer db.Close()

	// create repository instances
	customerRepository := storage.NewCustomerRepository(db)
	bookRepository := storage.NewBookRepository(db)
	orderRepository := storage.NewOrderRepository(db)

	// create security service
	securityService := &security.Service{}

	// create service instances
	customerService := customer.NewService(customerRepository, securityService)
	bookService := book.NewService(bookRepository)
	orderService := order.NewService(orderRepository, bookService)

	// Create the server instance
	server := server.New(customerService, bookService, orderService)

	// Start the server
	go func() {
		slog.Info(fmt.Sprintf("server is running on :%d", cfg.ServerPort))
		if err := server.E.Start(fmt.Sprintf(":%d", cfg.ServerPort)); err != nil {
			utils.LogErrorFatal(err)
		}
	}()

	// Wait for a signal to exit
	sig := <-c

	// Shutdown the server gracefully
	if err := server.E.Shutdown(ctx); err != nil {
		utils.LogErrorFatal(err)
	}

	slog.Info("Received signal, Server shut down gracefully", sig)
}
