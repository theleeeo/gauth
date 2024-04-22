package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/lerror"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/sdk"
	"github.com/theleeeo/thor/user"
)

func GenerateState() (string, error) {
	b := make([]byte, 32) // Adjust size as needed.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}

func (h *OAuthHandler) serveLogin(w http.ResponseWriter, r *http.Request, providerID string) error {
	provider, err := h.getProvider(providerID)
	if err != nil {
		return lerror.Wrap(err, "failed to get provider", http.StatusBadRequest)
	}

	// The error does not matter as a new session will be created either way.
	// We want to discard any old sessions anyways
	session, _ := h.store.New(r, h.sessionName)
	session.Values = make(map[interface{}]interface{})

	state, err := GenerateState()
	if err != nil {
		return lerror.Wrap(err, "failed to generate a state", http.StatusInternalServerError)
	}

	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		return lerror.Wrap(err, "failed to save the state", http.StatusInternalServerError)
	}

	returnTo, err := h.parseReturnTo(r)
	if err != nil {
		return err
	}

	if returnTo != "" {
		session.Values["return"] = returnTo
		if err := session.Save(r, w); err != nil {
			return lerror.Wrap(err, "failed to save the return url", http.StatusInternalServerError)
		}
	}

	redirectURL := fmt.Sprintf("%s/oauth/callback/%s/%s", h.appUrl.String(), provider.Type(), provider.Name())

	loginURL := provider.BuildLoginUrl(state, redirectURL)
	http.Redirect(w, r, loginURL, http.StatusFound)
	return nil
}

func (h *OAuthHandler) parseReturnTo(r *http.Request) (string, error) {
	returnTo := r.FormValue("return")
	if returnTo == "" {
		return "", nil
	}

	returnURL, err := url.Parse(returnTo)
	if err != nil {
		return "", lerror.Wrap(err, "failed to parse return url", http.StatusBadRequest)
	}

	if returnURL.Scheme == "" {
		return "", lerror.New("invalid return url: scheme is missing", http.StatusBadRequest)
	}

	if !slices.Contains(h.allowedReturns, returnURL.Host) {
		return "", lerror.New("invalid return url: host is not allowed", http.StatusBadRequest)
	}

	return returnTo, nil
}

func (h *OAuthHandler) serveCallback(w http.ResponseWriter, r *http.Request, providerID string) error {
	provider, err := h.getProvider(providerID)
	if err != nil {
		return lerror.Wrap(err, "failed to get provider", http.StatusBadRequest)
	}

	if err := r.ParseForm(); err != nil {
		return lerror.Wrap(err, "failed to parse form", http.StatusBadRequest)
	}

	state := r.FormValue("state")
	if state == "" {
		return lerror.New("state not found", http.StatusBadRequest)
	}

	session, err := h.store.New(r, h.sessionName)
	if err != nil {
		return lerror.Wrap(err, "failed to get session", http.StatusBadRequest)
	}

	if session.Values["state"] != state {
		return lerror.New("state mismatch", http.StatusBadRequest)
	}

	code := r.FormValue("code")
	if code == "" {
		return lerror.New("code not found", http.StatusBadRequest)
	}

	u, pr, err := provider.GetUser(code)
	if err != nil {
		return lerror.Wrap(err, "failed to get user from provider", http.StatusInternalServerError)
	}

	// ctx := sdk.WithClaims(r.Context(), &authorizer.Claims{Role: models.RoleAdmin})
	ctx := sdk.WithClaims(r.Context(), &authorizer.Claims{Permissions: map[string]string{"admin": "true"}})
	user, err := h.constructUser(ctx, u, pr)
	if err != nil {
		return err
	}

	token, err := h.auth.CreateToken(ctx, user)
	if err != nil {
		return lerror.Wrap(err, "failed to create token", http.StatusInternalServerError)
	}

	var returnTo string
	ret, ok := session.Values["return"]
	if ok {
		ret, ok := ret.(string)
		if ok {
			returnTo = ret
		}
	}

	cookie := &http.Cookie{
		Name:     h.cookieName,
		Domain:   returnTo,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   !(h.appUrl.Scheme == "http"), // If the app url is http, then the cookie is not secure. Default to secure in all other cases.
	}

	if returnTo == "" {
		returnTo = "/welcome.html"
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Location", returnTo)
	w.WriteHeader(http.StatusFound)
	return nil
}

// Try to get the user. If the user does not exist, create it.
func (h *OAuthHandler) constructUser(ctx context.Context, userModel models.User, provider models.UserProvider) (user.User, error) {
	// Try to get the u by the provider id
	u, err := h.app.GetUserByProviderID(ctx, provider.UserID)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return user.User{}, lerror.Wrap(err, "failed to get user", http.StatusInternalServerError)
	}

	// User was not found, check if it exist through another provider
	u, err = h.app.GetUserByEmail(ctx, u.Email)
	if err == nil {
		err = u.AddProvider(ctx, provider)
		if err != nil {
			return user.User{}, lerror.Wrap(err, "failed to add user provider", http.StatusInternalServerError)
		}
		return u, nil
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return user.User{}, lerror.Wrap(err, "failed to get user", http.StatusInternalServerError)
	}

	// User does not exist. Create the user
	u, err = h.app.CreateUser(ctx, userModel, provider)
	if err != nil {
		return user.User{}, lerror.Wrap(err, "failed to create user", http.StatusInternalServerError)
	}

	return u, nil
}
