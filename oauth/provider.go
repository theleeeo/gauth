package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/lerror"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/user"
)

type Provider interface {
	Type() string
	Name() string
	BuildLoginUrl(state, redirectUrl string) string
	GetUser(code string) (models.User, models.UserProvider, error)
}

type OAuthHandler struct {
	userService *user.Service
	auth        *authorizer.Authorizer
	store       *sessions.CookieStore

	providers []Provider

	appUrl      *url.URL
	cookieName  string
	sessionName string
	// What hosts are allowed to return to after login
	allowedReturns []*url.URL
}

func NewOAuthHandler(cfg *Config, userService *user.Service, auth *authorizer.Authorizer) (*OAuthHandler, error) {
	appUrl, err := url.Parse(cfg.AppURL)
	if err != nil {
		return nil, err
	}

	allowedReturns := make([]*url.URL, len(cfg.AllowedReturns))
	for i, u := range cfg.AllowedReturns {
		allowedReturns[i], err = url.Parse(u)
		if err != nil {
			return nil, err
		}
	}

	h := &OAuthHandler{
		userService:    userService,
		auth:           auth,
		store:          sessions.NewCookieStore([]byte(cfg.CookieSecret)),
		appUrl:         appUrl,
		cookieName:     cfg.CookieName,
		sessionName:    cfg.SessionName,
		allowedReturns: allowedReturns,
	}

	for _, providerCfg := range cfg.Providers {
		switch providerCfg.Type {
		case GithubProviderType:
			h.providers = append(h.providers, newGithub(providerCfg))
		case GoogleProviderType:
			h.providers = append(h.providers, newGoogle(providerCfg, cfg.AppURL))
		default:
			return nil, fmt.Errorf("unknown provider type: %s", providerCfg.Type)
		}
	}

	return h, nil
}

func (h *OAuthHandler) Register(mux *http.ServeMux) {
	mux.Handle("/oauth/", h)
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
	path := strings.TrimPrefix(r.URL.Path, "/oauth/")

	action, providerPath, ok := strings.Cut(path, "/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	var err error
	switch action {
	case "login":
		err = h.serveLogin(w, r, providerPath)
	case "callback":
		err = h.serveCallback(w, r, providerPath)
	default:
		http.NotFound(w, r)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), lerror.Status(err))
	}
}
