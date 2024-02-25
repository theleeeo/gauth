package runner

import (
	"net/http"
	"time"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/entrypoints"
)

type Runner struct {
	httpServer *http.Server
}

func New(cfg *Config) *Runner {
	auth := authorizer.New(cfg.SecretKey, cfg.ValidDuration)

	restAPI := entrypoints.NewRestHandler(auth)
	restAPI.Register(http.DefaultServeMux)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 8 * time.Second,
	}

	return &Runner{
		httpServer: httpServer,
	}
}

func (r *Runner) Run() error {
	return r.httpServer.ListenAndServe()
}
