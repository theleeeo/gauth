package runner

import (
	"net/http"
	"time"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/entrypoints"
	"github.com/theleeeo/thor/oauth"
)

type Runner struct {
	httpServer *http.Server
}

func New(cfg *Config) *Runner {
	auth := authorizer.New(cfg.AuthCfg.SecretKey, cfg.AuthCfg.ValidDuration)

	mux := http.DefaultServeMux

	restAPI := entrypoints.NewRestHandler(auth)
	restAPI.Register(mux)

	http.Handle("/", http.FileServer(http.Dir("public"))) // DEBUG ONLY THIS IS JUST WHEN DEVELOPING FOR TESTING

	if cfg.OAuthProviders.Github != nil {
		oauth := oauth.NewGithub(cfg.OAuthProviders.Github.ClientID, cfg.OAuthProviders.Github.ClientSecret)
		oauth.Register(mux)
	}

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
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
