package runner

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/entrypoints"
	"github.com/theleeeo/thor/oauth"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/user"
)

type Runner struct {
	httpServer *http.Server
}

func New(cfg *Config) (*Runner, error) {
	auth := authorizer.New(cfg.AuthCfg.SecretKey, cfg.AuthCfg.ValidDuration)

	repo, err := repo.New(cfg.RepoCfg)
	if err != nil {
		return nil, err
	}

	userSrv := user.NewService(repo)

	app := app.New(auth, userSrv)

	mux := http.DefaultServeMux

	restAPI := entrypoints.NewRestHandler(app)
	restAPI.Register(mux)

	http.Handle("/", http.FileServer(http.Dir("public"))) // DEBUG ONLY THIS IS JUST WHEN DEVELOPING FOR TESTING

	store := sessions.NewCookieStore([]byte(cfg.AuthCfg.SecretKey))

	for _, providerCfg := range cfg.OAuthProviders {
		switch providerCfg.Type {
		case oauth.GithubProviderType:
			oauth, _ := oauth.NewGithub(providerCfg, cfg.AppURL, store)
			oauth.Register(mux)
		}
	}

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 8 * time.Second,
	}

	return &Runner{
		httpServer: httpServer,
	}, nil
}

func (r *Runner) Run() error {
	return r.httpServer.ListenAndServe()
}
