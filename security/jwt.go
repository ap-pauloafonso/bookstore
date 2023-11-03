package security

import (
	"fmt"
	"github.com/ap-pauloafonso/bookstore/utils"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
)

var (
	jwtSecret = []byte("my-secret-key")
)

func GenerateJwtToken(email string, id int64) (string, error) {
	// Create a token with customer information
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// in case we want to hide the customerID from the end user we could use symmetric encryption here
		// but let's keep it simple, and just put the plain id there, as this information is not too sensitive.
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // 1 day expires
	})

	// Sign and get the complete encoded token as a string
	return token.SignedString(jwtSecret)

}

func JwtCheckMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the "Authorization" header from the request
			authHeader := c.Request().Header.Get("Authorization")

			// Check if the header is empty or doesn't start with "Bearer "
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			// Extract the token from the header (excluding "Bearer ")
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("invalid signing method")
				}
				return jwtSecret, nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Extract and store the email in the context
				if email, ok := claims["email"].(string); ok {
					c.Set("email", email)
				} else {
					return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
				}

				// Extract and store the id in the context
				if id, ok := claims["id"].(float64); ok {
					c.Set("id", int64(id))
				} else {
					return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
				}

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, utils.ErrorMessage{ErrorMessage: "Unauthorized"})
		}
	}
}
