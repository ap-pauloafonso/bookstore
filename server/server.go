package server

import (
	"errors"
	"fmt"
	"github.com/ap-pauloafonso/bookstore/book"
	"github.com/ap-pauloafonso/bookstore/customer"
	_ "github.com/ap-pauloafonso/bookstore/docs"
	"github.com/ap-pauloafonso/bookstore/order"
	"github.com/ap-pauloafonso/bookstore/security"
	"github.com/ap-pauloafonso/bookstore/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	echoSwagger "github.com/swaggo/echo-swagger"
	"log/slog"
	"net/http"
)

var (
	errInternalSever = errors.New("internal server error")
)

// Server represents the application instance
type Server struct {
	E               *echo.Echo
	customerService *customer.Service
	bookService     *book.Service
	orderService    *order.Service
}

type customerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ResultMessage struct {
	Message string `json:"message"`
}

// RegisterUserHandler
// @Summary customer Register
// @Description Register a new customer with email and password
// @Accept json
// @Produce json
// @Tags auth
// @Param user body customerRequest true "customer email/pass"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} utils.ErrorMessage
// @Failure 500 {object} utils.ErrorMessage
// @Router /api/register [post]
func (s *Server) RegisterUserHandler(c echo.Context) error {
	var u customerRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed to register customer: %s", err.Error())})
	}

	id, err := s.customerService.Register(c.Request().Context(), u.Email, u.Password)
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}

		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})
	}

	// generate jwt token
	tokenString, err := security.GenerateJwtToken(u.Email, *id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
}

// LoginUserHandler
// @Summary customer Login
// @Description Log in a customer with email and password
// @Accept json
// @Produce json
// @Tags auth
// @Param user body customerRequest true "customer email/pass"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} utils.ErrorMessage
// @Failure 500 {object} utils.ErrorMessage
// @Router /api/login [post]
func (s *Server) LoginUserHandler(c echo.Context) error {
	var u customerRequest

	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed login: %s", err.Error())})
	}

	newcustomer, err := s.customerService.Login(c.Request().Context(), u.Email, u.Password)
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}

		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})

	}

	// generate jwt token
	tokenString, err := security.GenerateJwtToken(u.Email, newcustomer.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "Internal Server Error"})
	}

	return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
}

// GetBooksHandler
// @Summary Get all books
// @Description Get a list of all books
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {array} book.Model
// @Failure 500 {object} utils.ErrorMessage
// @Router /api/books [get]
func (s *Server) GetBooksHandler(c echo.Context) error {

	books, err := s.bookService.GetAllBooks(c.Request().Context())
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})
	}

	return c.JSON(http.StatusOK, books)
}

// GetcustomerOrdersHandler
// @Summary Get customer orders
// @Description Get a list of orders for the authenticated customer
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Success 200 {array} order.Order
// @Failure 400 {object} utils.ErrorMessage
// @Failure 500 {object} utils.ErrorMessage
// @Router /api/orders [get]
func (s *Server) GetcustomerOrdersHandler(c echo.Context) error {
	ctx := c.Request().Context()
	email, ok := c.Get("email").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "email context value missing"})
	}

	customer, err := s.customerService.Getcustomer(ctx, email)
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})
	}
	orders, err := s.orderService.GetOrdersByCustomer(ctx, customer.Id)
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}

		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})
	}

	return c.JSON(http.StatusOK, orders)
}

// MakeOrderHandler
// @Summary Create an order
// @Description Create a new order with the provided items
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param orderItems body []order.OrderRequestItem true "List of order items"
// @Success 200 {object} order.Order
// @Failure 400 {object} utils.ErrorMessage
// @Failure 500 {object} utils.ErrorMessage
// @Router /api/orders [post]
func (s *Server) MakeOrderHandler(c echo.Context) error {
	orderItems := []order.OrderRequestItem{}

	if err := c.Bind(&orderItems); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: fmt.Sprintf("Failed login: %s", err.Error())})
	}

	ctx := c.Request().Context()
	customerID, ok := c.Get("id").(int64)
	if !ok {
		return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: "email context value missing"})
	}

	order, err := s.orderService.MakeOrder(ctx, customerID, orderItems)
	if err != nil {
		if utils.IsStorageRelatedError(err) {
			slog.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, utils.ErrorMessage{ErrorMessage: errInternalSever.Error()})
		}
		return c.JSON(http.StatusBadRequest, utils.ErrorMessage{ErrorMessage: err.Error()})
	}

	return c.JSON(http.StatusOK, order)

}

// New creates a new instance of the Server
func New(customerService *customer.Service, bookService *book.Service, orderService *order.Service) *Server {
	server := &Server{
		E:               echo.New(),
		customerService: customerService,
		bookService:     bookService,
		orderService:    orderService,
	}

	// set up API routes
	server.E.POST("/api/register", server.RegisterUserHandler)
	server.E.POST("/api/login", server.LoginUserHandler)
	server.E.GET("/api/books", server.GetBooksHandler)
	server.E.GET("/api/orders", server.GetcustomerOrdersHandler, security.JwtCheckMiddleware())
	server.E.POST("/api/orders", server.MakeOrderHandler, security.JwtCheckMiddleware())
	server.E.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// serve Swagger documentation on /swagger/index.html
	server.E.GET("/swagger/*", echoSwagger.WrapHandler)

	// set up middlewares
	server.E.Use(slogecho.New(slog.Default()))
	server.E.Use(middleware.Recover())

	return server

}
