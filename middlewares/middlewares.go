package middlewares

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"

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

func PrefixStripper(prefix string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = r.URL.Path[len(prefix):]
			h.ServeHTTP(w, r)
		})
	}
}

func ClaimsExtractor(publicKey []byte, cookieName string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := sdk.ExtractClaims(r, publicKey, cookieName)
			if err != nil {
				http.Error(w, "missing or invalid token", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = sdk.WithClaims(ctx, claims)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

// InternalErrorRedacter is a middleware that will redact internal error messages.
// It will replace the response body with a generic message and an id and log the original message.
func InternalErrorRedacter() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respCatcher := httptest.NewRecorder()
			h.ServeHTTP(respCatcher, r)

			if respCatcher.Code == http.StatusInternalServerError {
				responseId := rand.Intn(1000000)

				copyHeaders(w.Header(), respCatcher.Header())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("internal error, id: %d", responseId)))

				log.Printf("%d: %s: internal error: %s", responseId, r.URL.Path, respCatcher.Body.String())
				return
			}

			copyHeaders(w.Header(), respCatcher.Header())
			w.WriteHeader(respCatcher.Code)
			w.Write(respCatcher.Body.Bytes())
		})
	}
}

func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}
