package authorizer

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-jwt/jwt/v5"
)

type Authorizer struct {
	secret        []byte
	validDuration time.Duration

	parser *jwt.Parser
}

func New(secret string, validDuration time.Duration) *Authorizer {
	return &Authorizer{
		secret:        []byte(secret),
		validDuration: validDuration,
		parser:        jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired()),
	}
}

func (a *Authorizer) Decode(token string) (*Claims, error) {
	t, err := a.parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return &Claims{
		UserID:     claims["sub"].(string),
		Expiration: time.Unix(int64(claims["exp"].(float64)), 0),
		Role:       Role(claims["role"].(string)),
	}, nil
}

func (a *Authorizer) CreateToken(userID string, role Role) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  userID,
			"exp":  time.Now().Add(a.validDuration).Unix(),
			"role": role,
		})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
