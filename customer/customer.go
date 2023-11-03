package customer

import (
	"context"
	"errors"
	"net/mail"
	"time"
)

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errPasswordShort      = errors.New("invalid password: needs to have at least 3 characters")
	errPasswordLong       = errors.New("invalid password: exceed the max amount of 50 characters")
	errEmailLong          = errors.New("invalid email: exceed the max amount of 255 characters")
	errEmailInvalid       = errors.New("invalid email")
	errEmailAlreadyTaken  = errors.New("email already registered")
	errStoringcustomer    = errors.New("error storing customer")
	errcustomerNotFound   = errors.New("error customer not found")
)

type Service struct {
	repository Repository
	security   SecurityService
}

func NewService(customerRepository Repository, securityService SecurityService) *Service {
	return &Service{customerRepository, securityService}
}

type Model struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Repository interface {
	SaveCustomer(ctx context.Context, email, password string, createdAt time.Time) (*int64, error)
	GetCustomer(ctx context.Context, email string) (*Model, error)
}

type SecurityService interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (s *Service) Register(ctx context.Context, email, password string) (*int64, error) {
	if len(password) < 3 {
		return nil, errPasswordShort
	}

	if len(password) > 50 {
		return nil, errPasswordLong
	}

	if len(email) > 255 {
		return nil, errEmailLong
	}

	if !isValidEmail(email) {
		return nil, errEmailInvalid
	}

	_, err := s.repository.GetCustomer(ctx, email)
	if err == nil {
		return nil, errEmailAlreadyTaken
	}

	hashedPassword, err := s.security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	id, err := s.repository.SaveCustomer(ctx, email, hashedPassword, time.Now())
	if err != nil {
		return nil, errStoringcustomer
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*Model, error) {
	customer, err := s.repository.GetCustomer(ctx, email)
	if err != nil {
		return nil, errInvalidCredentials
	}

	if !s.security.CheckPasswordHash(password, customer.Password) {
		return nil, errInvalidCredentials
	}

	return customer, nil
}

func (s *Service) Getcustomer(ctx context.Context, email string) (*Model, error) {
	customer, err := s.repository.GetCustomer(ctx, email)
	if err != nil {
		return nil, errcustomerNotFound
	}

	return customer, nil
}
