package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/theleeeo/thor/authorizer"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
	"github.com/theleeeo/thor/sdk"
)

func GenerateState() (string, error) {
	b := make([]byte, 32) // Adjust size as needed.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return state, nil
}

func (h *OAuthHandler) serveLogin(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := h.getProvider(providerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// The error does not matter as a new session will be created either way.
	// We want to discard any old sessions anyways
	session, _ := h.store.New(r, h.sessionName)

	state, err := GenerateState()
	if err != nil {
		http.Error(w, fmt.Errorf("failed to generate a state: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, fmt.Errorf("failed to save the state: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	returnTo := r.FormValue("return")
	if returnTo != "" {
		returnURL, err := url.Parse(returnTo)
		if err != nil {
			http.Error(w, fmt.Errorf("failed to parse return url: %w", err).Error(), http.StatusBadRequest)
			return
		}

		if returnURL.Scheme == "" {
			http.Error(w, "invalid return url: scheme is missing", http.StatusBadRequest)
			return
		}

		if !slices.Contains(h.allowedReturns, returnURL.Host) {
			http.Error(w, "invalid return url: host is not allowed", http.StatusBadRequest)
			return
		}

		session.Values["return"] = returnTo
		if err := session.Save(r, w); err != nil {
			http.Error(w, fmt.Errorf("failed to save the return url: %w", err).Error(), http.StatusInternalServerError)
			return
		}
	}

	redirectURL := fmt.Sprintf("%s/oauth/callback/%s/%s", h.appUrl.String(), provider.Type(), provider.Name())

	loginURL := provider.BuildLoginUrl(state, redirectURL)
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *OAuthHandler) serveCallback(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := h.getProvider(providerID)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get provider: %w", err).Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Errorf("failed to parse form: %w", err).Error(), http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")
	if state == "" {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	session, err := h.store.New(r, h.sessionName)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get session: %w", err).Error(), http.StatusBadRequest)
		return
	}

	if session.Values["state"] != state {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	u, err := provider.GetUser(code)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get user from provider: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var user *models.User
	ctx := sdk.WithClaims(r.Context(), &authorizer.Claims{Role: models.RoleAdmin})
	user, err = h.app.GetUserByProviderID(ctx, u.Providers[0].UserID)
	if err != nil {
		if !errors.Is(err, repo.ErrNotFound) {
			http.Error(w, fmt.Errorf("failed to get user: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		// User was not found, check if it exist through another provider
		user, err = h.app.GetUserByEmail(ctx, u.Email)
		if err != nil {
			if !errors.Is(err, repo.ErrNotFound) {
				http.Error(w, fmt.Errorf("failed to get user: %w", err).Error(), http.StatusInternalServerError)
				return
			}

			// User does not exist. Create the user
			user, err = h.app.CreateUser(ctx, u)
			if err != nil {
				http.Error(w, fmt.Errorf("failed to create user: %w", err).Error(), http.StatusInternalServerError)
				return
			}
		} else {
			err = h.app.AddUserProvider(ctx, user.ID, u.Providers[0])
			if err != nil {
				http.Error(w, fmt.Errorf("failed to add user provider: %w", err).Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	token, err := h.app.CreateToken(ctx, user)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     h.cookieName,
		Domain:   h.appUrl.Hostname(), // TODO: THIS Will have to change when redirects are used
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   !(h.appUrl.Scheme == "http"), // If the app url is http, then the cookie is not secure. Default to secure in all other cases.
	}

	returnTo := session.Values["return"]
	if returnTo == nil {
		returnTo = "/welcome.html"
	}

	http.SetCookie(w, cookie)
	fmt.Println("Redirecting to:", returnTo.(string))
	w.Header().Set("Location", returnTo.(string))
	w.WriteHeader(http.StatusFound)
}
