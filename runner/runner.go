package runner

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/entrypoints"
	"github.com/theleeeo/thor/middlewares"
	"github.com/theleeeo/thor/oauth"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/role"
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

	auth, err := authorizer.New(&authorizer.Config{
		AppUrl:        cfg.AppUrl,
		PrivateKey:    privKey,
		PublicKey:     pubKey,
		ValidDuration: cfg.AuthCfg.ValidDuration,
	})
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

	// User service
	//
	userSrv := user.NewService(repo)

	//
	// Role service
	//
	roleSrv := role.NewService(repo)

	//
	// App
	//
	appImpl := app.New(auth, userSrv, roleSrv)

	rootMux := http.DefaultServeMux

	//
	// Rest handler
	//
	restAPI := entrypoints.NewRestHandler(appImpl, cfg.OAuthConfig.CookieName)

	apiMux := http.NewServeMux()
	restAPI.Register(apiMux)
	rootMux.Handle("/api/", middlewares.Chain(apiMux, middlewares.ClaimsExtractor(auth.PublicKey(), cfg.OAuthConfig.CookieName), middlewares.PrefixStripper("/api")))
	// rootMux.Handle("/api/", middlewares.Chain(apiMux, middlewares.PrefixStripper("/api")))

	errorPageDirector, err := middlewares.ErrorPageDirector(map[int]string{
		404: "404.html",
		400: "400.html",
	}, "internal.html")
	if err != nil {
		return err
	}

	http.Handle("/", errorPageDirector(http.FileServer(HTMLDir("public"))))

	//
	//	Create the oauth handler
	//
	if cfg.OAuthConfig.AppURL == "" {
		cfg.OAuthConfig.AppURL = cfg.AppUrl
	}

	oauthHandler, err := oauth.NewOAuthHandler(cfg.OAuthConfig, userSrv, auth)
	if err != nil {
		return err
	}

	rootMux.Handle("/oauth/", errorPageDirector(oauthHandler))

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      middlewares.Chain(rootMux, middlewares.InternalErrorRedacter()),
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 8 * time.Second,
	}

	r := &Runner{
		httpServer: httpServer,
	}

	return r.httpServer.ListenAndServe()
}

type HTMLDir string

func (d HTMLDir) Open(name string) (http.File, error) {
	if name[len(name)-1] != '/' && filepath.Ext(name) == "" {
		name += ".html"
	}

	return http.Dir(d).Open(name)
}
