package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/theleeeo/thor/models"
)

const (
	googleLoginEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint = "https://oauth2.googleapis.com/token"
)

type googleHandler struct {
	clientID     string
	clientSecret string
	name         string
	appBaseURL   string
}

func newGoogle(cfg ProviderConfig, appBaseURL string) *googleHandler {
	return &googleHandler{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		name:         cfg.Name,
		appBaseURL:   appBaseURL,
	}
}

func (g *googleHandler) BuildLoginUrl(state, redirectURL string) string {
	// %20 is encoded as a space in the URL
	scope := "openid%20email%20profile"

	return fmt.Sprintf("%s?response_type=code&scope=%s&client_id=%s&state=%s&redirect_uri=%s", googleLoginEndpoint, scope, g.clientID, state, redirectURL)
}

func (g *googleHandler) Name() string {
	return g.name
}

func (g *googleHandler) Type() string {
	return string(GoogleProviderType)
}

func (g *googleHandler) GetUser(code string) (*models.User, error) {
	token, err := g.getIdToken(code)
	if err != nil {
		return nil, err
	}

	// It is safe to do this unverified because we know that the token is directly from google
	t, _, err := jwt.NewParser().ParseUnverified(token, &googleClaims{})
	if err != nil {
		return nil, fmt.Errorf("could not parse JWT token: %v", err)
	}

	claims, ok := t.Claims.(*googleClaims)
	if !ok {
		return nil, fmt.Errorf("could not parse JWT claims")
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("email not verified")
	}

	return &models.User{
		Name:  claims.Given_name + " " + claims.Family_name,
		Email: claims.Email,
		Providers: []models.UserProvider{
			{
				UserID: claims.Subject,
				Type:   models.UserProviderTypeGoogle,
			},
		},
	}, nil
}

func (g *googleHandler) getIdToken(code string) (string, error) {
	redirectURL := fmt.Sprintf("%s/oauth/callback/google/%s", g.appBaseURL, g.name)
	reqURL := fmt.Sprintf("%s?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s", googleTokenEndpoint, g.clientID, g.clientSecret, code, redirectURL)
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
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("non-ok status code: %d, %s", res.StatusCode, body)
	}

	defer res.Body.Close()

	var respBody = struct {
		// AccessToken string `json:"access_token"`
		IdToken string `json:"id_token"`
		// GrantType   string `json:"grant_type"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("could not parse JSON response: %v", err)
	}

	return respBody.IdToken, nil
}
