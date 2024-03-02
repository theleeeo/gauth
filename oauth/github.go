package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theleeeo/thor/models"
)

const (
	cookieName     = "thor_token"
	githubLoginUrl = "https://github.com/login/oauth/authorize?client_id=%s&state=%s&redirect_uri=%s"
)

type githubHandler struct {
	clientID     string
	clientSecret string
	name         string
}

func newGithub(cfg ProviderConfig) *githubHandler {
	return &githubHandler{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		name:         cfg.Name,
	}
}

func extractToken(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (g *githubHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("GET /oauth/whoami/github/%s", g.name), func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		req.Header.Set("Authorization", fmt.Sprintf("token %s", extractToken(r)))
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

func (g *githubHandler) BuildLoginUrl(state, redirectURL string) string {
	return fmt.Sprintf(githubLoginUrl, g.clientID, state, redirectURL)
}

func (g *githubHandler) Name() string {
	return g.name
}

func (g *githubHandler) Type() string {
	return string(GithubProviderType)
}

func (g *githubHandler) GetUser(code string) (*models.User, error) {
	token, err := g.getAccessToken(code)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var user = struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &models.User{
		Nickname: user.Login,
		Provider: models.UserProvider{
			UserID: fmt.Sprintf("%d", user.ID),
			Type:   models.UserProviderTypeGithub,
		},
	}, nil
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
