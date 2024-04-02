package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/models"
)

type Provider interface {
	Type() string
	Name() string
	BuildLoginUrl(state, redirectUrl string) string
	GetUser(code string) (*models.User, error)
}

type OAuthHandler struct {
	appUrl    *url.URL
	providers []Provider
	store     *sessions.CookieStore
	app       *app.App
}

func NewOAuthHandler(cfg *Config, app *app.App) (*OAuthHandler, error) {
	appUrl, err := url.Parse(cfg.AppURL)
	if err != nil {
		return nil, err
	}

	h := &OAuthHandler{
		appUrl: appUrl,
		app:    app,
		store:  sessions.NewCookieStore([]byte(cfg.CookieSecret)),
	}

	for _, providerCfg := range cfg.Providers {
		switch providerCfg.Type {
		case GithubProviderType:
			h.providers = append(h.providers, newGithub(providerCfg))
		default:
			return nil, fmt.Errorf("unknown provider type: %s", providerCfg.Type)
		}
	}

	return h, nil
}

func (h *OAuthHandler) getProvider(path string) (Provider, error) {
	for _, p := range h.providers {
		if fmt.Sprintf("%s/%s", p.Type(), p.Name()) == path {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no provider found for path: %s", path)
}

func (h *OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Trim the leading slash
	path := r.URL.Path[1:]

	action, providerPath, ok := strings.Cut(path, "/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch action {
	case "login":
		h.serveLogin(w, r, providerPath)
	case "callback":
		h.serveCallback(w, r, providerPath)
	default:
		http.NotFound(w, r)
		return
	}
}
