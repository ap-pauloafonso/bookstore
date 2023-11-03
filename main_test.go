package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	networkName := "new-network"
	newNetwork, err := testcontainers.GenericNetwork(context.Background(), testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to start network: %v", err)

	}
	defer newNetwork.Remove(context.Background())

	reqPostgresqlContainer := testcontainers.ContainerRequest{
		NetworkAliases: map[string][]string{networkName: {"postgres"}},
		Networks:       []string{networkName},
		Image:          "postgres:latest",
		ExposedPorts:   []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "MY_DB",
		},
		WaitingFor: wait.ForAll(wait.ForListeningPort("5432/tcp"), wait.ForLog("database system is ready to accept connections")),
	}
	postgresC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: reqPostgresqlContainer,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	defer postgresC.Terminate(context.Background())

	time.Sleep(3 * time.Second)
	reqServerContainer := testcontainers.ContainerRequest{
		Networks: []string{networkName},
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "./Dockerfile",
		},
		Env: map[string]string{
			"SERVER_PORT":         "8080",
			"POSTGRES_CONNECTION": fmt.Sprintf("host=%s port=%s user=postgres password=test dbname=MY_DB sslmode=disable", "postgres", "5432"),
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("http server started on [::]:8080"),
	}

	serverC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: reqServerContainer,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start server container: %v", err)
	}
	defer serverC.Terminate(context.Background())

	host, err := serverC.Host(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	port, err := serverC.MappedPort(context.Background(), "8080")
	if err != nil {
		t.Error(err)
	}

	url := fmt.Sprintf("http://%s:%s", host, port.Port())

	t.Run("health test", func(t *testing.T) {
		resp, err := http.Get(url + "/health")
		if err != nil {
			t.Fatal("health endpoint shouldn't fail", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("api health should return 200, it returned: ", resp.StatusCode)
		}
	})

	type Model struct {
		ID     int64   `json:"id"`
		Title  string  `json:"title"`
		Author string  `json:"author"`
		Price  float64 `json:"price"`
	}

	var booksResult []Model
	t.Run("api get books", func(t *testing.T) {
		resp, err := http.Get(url + "/api/books")
		if err != nil {
			t.Fatal("books endpoints shouldn't fail", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("api health should return 200, it returned: ", resp.StatusCode)
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("Failed to read response body", err)
		}

		var responseSlice []Model

		// Unmarshal the response JSON into the slice of Model
		if err := json.Unmarshal(responseBody, &responseSlice); err != nil {
			t.Fatal("Failed to unmarshal response JSON", err)
		}

		if len(responseSlice) < 2 {
			t.Fatal("should have more than 2, because there is a migration")
		}

		booksResult = responseSlice

	})

	// register customer

	t.Run("api register", func(t *testing.T) {
		jsonData, err := json.Marshal(map[string]string{"email": "paulo@gmail.com", "password": "123"})
		if err != nil {
			t.Fatal("JSON serialization error", err)
		}

		respRegister, err := http.Post(url+"/api/register", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal("get shouldn't fail", err)
		}
		defer respRegister.Body.Close()
		if respRegister.StatusCode != http.StatusOK {
			t.Fatal("register should return 200, it returned: ", respRegister.StatusCode)
		}

		responseBody, err := io.ReadAll(respRegister.Body)
		if err != nil {
			t.Fatal("Failed to read response body", err)
		}

		var responseStruct struct {
			Token string `json:"token"`
		}

		if err := json.Unmarshal(responseBody, &responseStruct); err != nil {
			t.Fatal("Failed to unmarshal response JSON", err)

		}

	})

	var token string
	t.Run("api login", func(t *testing.T) {
		jsonData, err := json.Marshal(map[string]string{"email": "paulo@gmail.com", "password": "123"})
		if err != nil {
			t.Fatal("JSON serialization error", err)
		}

		respLogin, err := http.Post(url+"/api/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal("get shouldn't fail", err)
		}
		defer respLogin.Body.Close()
		if respLogin.StatusCode != http.StatusOK {
			t.Fatal("login should return 200, it returned: ", respLogin.StatusCode)
		}

		responseBody, err := io.ReadAll(respLogin.Body)
		if err != nil {
			t.Fatal("Failed to read response body", err)
		}

		var responseStruct struct {
			Token string `json:"token"`
		}

		if err := json.Unmarshal(responseBody, &responseStruct); err != nil {
			t.Fatal("Failed to unmarshal response JSON", err)

		}

		token = responseStruct.Token
	})

	type OrderRequestItem struct {
		BookID   int64 `json:"book_id"`
		Quantity int   `json:"quantity"`
	}

	b1 := booksResult[0]
	b2 := booksResult[1]

	t.Run("api create order", func(t *testing.T) {

		jsonData, err := json.Marshal([]OrderRequestItem{
			{
				BookID:   b1.ID,
				Quantity: 1,
			},
			{
				BookID:   b2.ID,
				Quantity: 3,
			},
		})
		if err != nil {
			t.Fatal("JSON serialization error", err)
		}

		req, err := http.NewRequest("POST", url+"/api/orders", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatal("Failed to create create order request", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("post order should return 200, it returned: ", resp.StatusCode)
		}

		type OrderItem struct {
			BookID   int64   `json:"book_id"`
			Quantity int     `json:"quantity"`
			Price    float64 `json:"price"`
		}

		var order struct {
			ID        int64       `json:"id"`
			Total     float64     `json:"total"`
			OrderDate time.Time   `json:"order_date"`
			Items     []OrderItem `json:"items"`
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("Failed to read response body", err)
		}

		if err := json.Unmarshal(responseBody, &order); err != nil {
			t.Fatal("Failed to unmarshal response JSON", err)
		}

		if order.ID == 0 {
			t.Fatal("order should have a valid id", err)
		}

		if len(order.Items) != 2 {
			t.Fatal("order should have 2 item", err)
		}

		expectedTotal := (b1.Price * 1) + (b2.Price * 3)

		if order.Total != expectedTotal {
			t.Fatal("expected total: ", expectedTotal, " got: ", order.Total)
		}

	})

	// get order
	t.Run("api get orders", func(t *testing.T) {

		req, err := http.NewRequest("GET", url+"/api/orders", nil)
		if err != nil {
			t.Fatal("Failed to create get orders request", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("get order should return 200, it returned: ", resp.StatusCode)
		}

		type OrderItem struct {
			BookID   int64   `json:"book_id"`
			Quantity int     `json:"quantity"`
			Price    float64 `json:"price"`
		}

		var orders []struct {
			ID        int64       `json:"id"`
			Total     float64     `json:"total"`
			OrderDate time.Time   `json:"order_date"`
			Items     []OrderItem `json:"items"`
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal("Failed to read response body", err)
		}

		if err := json.Unmarshal(responseBody, &orders); err != nil {
			t.Fatal("Failed to unmarshal response JSON", err)
		}

		if len(orders) == 0 {
			t.Fatal("expected 1 order but got zero")
		}

		order := orders[0] // get the first order

		if order.ID == 0 {
			t.Fatal("order should have a valid id", err)
		}

		if len(order.Items) != 2 {
			t.Fatal("order should have 2 item", err)
		}

		expectedTotal := (b1.Price * 1) + (b2.Price * 3)

		if order.Total != expectedTotal {
			t.Fatal("expected total: ", expectedTotal, " got: ", order.Total)
		}

	})
}
