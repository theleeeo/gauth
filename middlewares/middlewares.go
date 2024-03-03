package middlewares

import (
	"net/http"

	"github.com/theleeeo/thor/sdk"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func ChainFunc(h http.HandlerFunc, m ...Middleware) http.Handler {
	return Chain(http.HandlerFunc(h), m...)
}

func WithMiddleware(h http.Handler, m Middleware) http.Handler {
	return m(h)
}

func ClaimsExtractor(publicKey []byte) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := sdk.ExtractClaims(r, publicKey)
			if err == nil {
				// If there is no error, we will add the claims to the context
				ctx := r.Context()
				ctx = sdk.WithClaims(ctx, claims)
				r = r.WithContext(ctx)
			}

			h.ServeHTTP(w, r)
		})
	}
}
