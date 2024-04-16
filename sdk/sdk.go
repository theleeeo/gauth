package sdk

import (
	"context"
	"net/http"

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

func ExtractClaims(r *http.Request, publicKey []byte, cookieName string) (*authorizer.Claims, error) {
	token, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	t, err := jwt.ParseWithClaims(token.Value, &authorizer.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseEdPublicKeyFromPEM(publicKey)
	})
	if err != nil {
		return nil, err
	}

	claims := t.Claims.(*authorizer.Claims)

	return claims, nil
}

// func UserIsRole(ctx context.Context, role models.Role) bool {
// 	claims := ClaimFromCtx(ctx)
// 	if claims == nil {
// 		return false
// 	}
// 	return claims.Role == role
// }

func UserIs(ctx context.Context, userID string) bool {
	claims := ClaimFromCtx(ctx)
	if claims == nil {
		return false
	}
	return claims.UserID == userID
}
