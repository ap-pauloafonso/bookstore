package security

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Service struct{}

var (
	errHashingPassword = errors.New("problem hashing password")
)

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error(err.Error())
		return "", errHashingPassword
	}

	return string(hash), nil
}

func (s *Service) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
