package runner

import (
	"net/http"
	"os"
	"time"

	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/entrypoints"
	"github.com/theleeeo/thor/middlewares"
	"github.com/theleeeo/thor/oauth"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/user"
)

type Runner struct {
	httpServer *http.Server
}

func Run(cfg *Config) error {
	//
	// Create the authorizer
	//
	privKey, err := os.ReadFile(cfg.AuthCfg.PrivateKey)
	if err != nil {
		return err
	}

	pubKey, err := os.ReadFile(cfg.AuthCfg.PublicKey)
	if err != nil {
		return err
	}

	auth, err := authorizer.New(privKey, pubKey, cfg.AuthCfg.ValidDuration)
	if err != nil {
		return err
	}

	//
	// Create the repository
	//
	repo, err := repo.New(cfg.RepoCfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	// if err := repo.Migrate(); err != nil {
	// 	return err
	// }

	if err := repo.Ping(); err != nil {
		return err
	}

	//
	// Create the user service
	//
	userSrv := user.NewService(repo)

	//
	// Create the app
	//
	appImpl := app.New(auth, userSrv)

	mux := http.DefaultServeMux

	restAPI := entrypoints.NewRestHandler(appImpl)
	restAPI.Register(mux)

	http.Handle("/", http.FileServer(http.Dir("public"))) // DEBUG ONLY THIS IS JUST WHEN DEVELOPING FOR TESTING

	//
	//	Create the oauth handler
	//
	oauthHandler, err := oauth.NewOAuthHandler(cfg.OAuthConfig, appImpl)
	if err != nil {
		return err
	}
	oauthHandler.Register(mux)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      middlewares.Chain(mux, middlewares.InternalErrorRedacter(), middlewares.ClaimsExtractor(auth.PublicKey())),
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 8 * time.Second,
	}

	r := &Runner{
		httpServer: httpServer,
	}

	return r.httpServer.ListenAndServe()
}
