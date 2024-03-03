package authorizer

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var _ jwt.Claims = &Claims{}

type Claims struct {
	// jwt.RegisteredClaims
	UserID     string    `json:"sub"`
	Expiration time.Time `json:"exp"`
	Role       Role      `json:"role"`
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.Expiration), nil
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetIssuer() (string, error) {
	return "", nil
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetSubject() (string, error) {
	return c.UserID, nil
}

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)
