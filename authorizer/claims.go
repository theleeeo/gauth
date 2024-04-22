package authorizer

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Issuer      string            `json:"iss"`
	UserID      string            `json:"sub"`
	ExpiresAt   time.Time         `json:"exp"`
	Permissions map[string]string `json:"perms"`
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(c.ExpiresAt), nil
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetIssuer() (string, error) {
	return c.Issuer, nil
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *Claims) GetSubject() (string, error) {
	return c.UserID, nil
}
