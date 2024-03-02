package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
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

	session, err := h.store.New(r, "thor_session")
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

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

	fmt.Println("loginnn-state: ", state)

	redirectURL := fmt.Sprintf("%s/oauth/callback/%s/%s", h.appUrl.String(), provider.Type(), provider.Name())
	loginURL := provider.BuildLoginUrl(state, redirectURL)
	http.Redirect(w, r, loginURL, http.StatusFound)
}

func (h *OAuthHandler) validateState(r *http.Request, w http.ResponseWriter) (bool, error) {
	state := r.FormValue("state")
	if state == "" {
		return false, nil
	}

	session, err := h.store.Get(r, "thor_session")
	if err != nil {
		return false, fmt.Errorf("failed to get session: %w", err)
	}

	defer func() {
		// Clear the state. It is not needed anymore after the oauth flow is complete.
		session.Values["state"] = nil
		if err := session.Save(r, w); err != nil {
			log.Printf("failed to remove state: %v", err)
			return
		}
	}()

	fmt.Println("session-state: ", session.Values["state"])
	fmt.Println("callbac-state: ", state)

	if session.Values["state"] == state {
		return true, nil
	}

	return false, nil
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

	ok, err := h.validateState(r, w)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to validate state: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	user, err := provider.GetUser(code)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to get user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "thor_token",
		Domain:   h.appUrl.Hostname(), // THIS Will have to change when redirects are used
		Value:    user.Nickname,       // This should be the jwt
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   !(h.appUrl.Scheme == "http"), // If the app url is http, then the cookie is not secure. Default to secure in all other cases.
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Location", "/welcome.html")
	w.WriteHeader(http.StatusFound)
}
