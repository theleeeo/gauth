package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
)

const (
	githubLoginUrl = "https://github.com/login/oauth/authorize?client_id=%s&state=%s&redirect_uri=%s"
)

type githubHandler struct {
	appUrl       *url.URL
	clientID     string
	clientSecret string
	name         string

	store *sessions.CookieStore
}

func NewGithub(cfg ProviderConfig, appUrl string, store *sessions.CookieStore) (*githubHandler, error) {
	url, err := url.Parse(appUrl)
	if err != nil {
		return nil, fmt.Errorf("could not parse app url: %v", err)
	}

	return &githubHandler{
		appUrl:       url,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		name:         cfg.Name,
		store:        store,
	}, nil
}

func GenerateState() (string, error) {
	b := make([]byte, 32) // Adjust size as needed.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.StdEncoding.EncodeToString(b)
	return state, nil
}

func (g *githubHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("GET /oauth/callback/github/%s", g.name), g.redirectCallback)
	mux.HandleFunc(fmt.Sprintf("GET /oauth/login/github/%s", g.name), g.login)
	mux.HandleFunc(fmt.Sprintf("GET /oauth/whoami/github/%s", g.name), func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("cookie: ", r.Cookies()[8].Value)
		req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		req.Header.Set("Authorization", fmt.Sprintf("token %s", r.Cookies()[8].Value))
		req.Header.Add("Accept", "application/vnd.github.v3+json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		fmt.Println("status: ", res.Status)

		var user = struct {
			Login      string `json:"login"`
			ID         int    `json:"id"`
			Avatar_url string `json:"avatar_url"`
		}{}

		if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	})
}

func (g *githubHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := GenerateState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := g.store.New(r, "thor-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("loginnn-state: ", state)

	redirectURL := fmt.Sprintf("%s/oauth/callback/github/%s", g.appUrl, g.name)
	http.Redirect(w, r, fmt.Sprintf(githubLoginUrl, g.clientID, state, redirectURL), http.StatusFound)
}

func (g *githubHandler) redirectCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")
	if state == "" {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	session, err := g.store.Get(r, "thor-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("session-state: ", session.Values["state"])
	fmt.Println("callbac-state: ", state)

	if session.Values["state"] != state {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	token, err := g.getAccessToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Clear the state. It is not needed anymore after the oauth flow is complete.
	session.Values["state"] = nil
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "thor_token",
		Domain:   g.appUrl.Hostname(),
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   !(g.appUrl.Scheme == "http"), // If the app url is http, then the cookie is not secure. Default to secure in all other cases.
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Location", "/welcome.html")
	w.WriteHeader(http.StatusFound)
}

func (g *githubHandler) getAccessToken(code string) (string, error) {
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", g.clientID, g.clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("could not create HTTP request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not send HTTP request: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-ok status code: %d", res.StatusCode)
	}

	defer res.Body.Close()

	// Parse the request body
	var respBody = struct {
		AccessToken string `json:"access_token"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("could not parse JSON response: %v", err)
	}

	return respBody.AccessToken, nil
}
