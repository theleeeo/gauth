package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theleeeo/thor/models"
)

const (
	githubLoginEndpoint = "https://github.com/login/oauth/authorize"
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

func (g *githubHandler) BuildLoginUrl(state, redirectURL string) string {
	scopes := "user:email%20read:user"
	return fmt.Sprintf("%s?client_id=%s&state=%s&redirect_uri=%s&scope=%s", githubLoginEndpoint, g.clientID, state, redirectURL, scopes)
}

func (g *githubHandler) Name() string {
	return g.name
}

func (g *githubHandler) Type() string {
	return string(GithubProviderType)
}

func (g *githubHandler) GetUser(code string) (models.User, models.UserProvider, error) {
	token, err := g.getAccessToken(code)
	if err != nil {
		return models.User{}, models.UserProvider{}, err
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.User{}, models.UserProvider{}, err
	}
	defer res.Body.Close()

	var user = struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return models.User{}, models.UserProvider{}, err
	}

	return models.User{
			Name:  user.Name,
			Email: user.Email,
		}, models.UserProvider{
			UserID: fmt.Sprintf("%d", user.ID),
			Type:   models.UserProviderTypeGithub,
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
