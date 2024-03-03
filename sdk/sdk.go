package sdk

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theleeeo/thor/authorizer"
)

type ClaimsContextKey string

func ClaimFromCtx(ctx context.Context) *authorizer.Claims {
	claims, ok := ctx.Value(ClaimsContextKey("claims")).(*authorizer.Claims)
	if !ok {
		return nil
	}
	return claims
}

func WithClaims(ctx context.Context, claims *authorizer.Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey("claims"), claims)
}

func ExtractClaims(r *http.Request, publicKey []byte) (*authorizer.Claims, error) {
	token, err := r.Cookie("thor_token")
	if err != nil {
		return nil, err
	}

	t, err := jwt.Parse(token.Value, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseEdPublicKeyFromPEM(publicKey)
	})
	if err != nil {
		return nil, err
	}

	claims := t.Claims.(jwt.MapClaims)

	return &authorizer.Claims{
		UserID:     claims["sub"].(string),
		Expiration: time.Unix(int64(claims["exp"].(float64)), 0),
		Role:       authorizer.Role(claims["role"].(string)),
	}, nil
}
