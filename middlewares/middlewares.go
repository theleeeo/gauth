package middlewares

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/theleeeo/thor/sdk"
)

const errorpagesDirectory = "errorpages"

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func ChainFunc(h http.HandlerFunc, m ...Middleware) http.Handler {
	return Chain(h, m...)
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

// ErrorPageDirector is a middleware that will serve error pages based on the response status code.
// It will look for specific error pages based on the status code and if it does not find one, it will use the catchall error page.
// The error pages provided should be in a directory called "errorpages" and the paths provided should be relative to that directory.
//
// The error pages can be go-templates and the following data will be passed to them:
// - Code: the status code of the response
// - Message: the response body
// - ID: a random id that can be used to track the error
func ErrorPageDirector(errorPages map[int]string, catchall string) (Middleware, error) {
	var filepaths []string
	for _, v := range errorPages {
		filepaths = append(filepaths, filepath.Join(errorpagesDirectory, v))
	}

	filepaths = append(filepaths, filepath.Join(errorpagesDirectory, catchall))

	// Load the HTML template from the errorpages directory
	tmpl, err := template.ParseFiles(filepaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load error pages: %w", err)
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			respCatcher := httptest.NewRecorder()
			h.ServeHTTP(respCatcher, r)

			// If the response is successful, just return it
			if respCatcher.Code < 400 {
				copyHeaders(w.Header(), respCatcher.Header())
				w.WriteHeader(respCatcher.Code)
				w.Write(respCatcher.Body.Bytes())
				return
			}

			errorData := struct {
				Code    int
				Message string
				ID      int
			}{
				Code:    respCatcher.Code,
				Message: respCatcher.Body.String(),
				ID:      rand.Intn(1000000),
			}

			slog.Error("error serving request", "id", errorData.ID, "code", errorData.Code, "message", errorData.Message, "url", r.URL.Path)

			// If the response is an error, check if there is a specific error page for it
			if pageName, ok := errorPages[respCatcher.Code]; ok {
				w.WriteHeader(respCatcher.Code)
				err := tmpl.ExecuteTemplate(w, pageName, errorData)
				if err != nil {
					// Since the template is written directly to the response writer it will still be partially written to the client, even if there is an error halfway through.
					// Just log it and accept the incomplete error page.
					// Write a working template next time!
					slog.Error("failed to render error page", "err", err)
					return
				}
				return
			}

			// If there is no specific error page, use the catchall error page
			w.WriteHeader(respCatcher.Code)
			err := tmpl.ExecuteTemplate(w, catchall, errorData)
			if err != nil {
				slog.Error("failed to render error page", "err", err)
				return
			}
		})
	}, nil
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
