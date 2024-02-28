package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type githubHandler struct {
	clientID     string
	clientSecret string
}

func NewGithub(clientID, clientSecret string) *githubHandler {
	return &githubHandler{
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (g *githubHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /oauth/callback/github", g.redirectCallback)
}

func (g *githubHandler) redirectCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", g.clientID, g.clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stdout, "could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stdout, "unexpected status code: %d", res.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	// Parse the request body
	var respBody = struct {
		AccessToken string `json:"access_token"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
		fmt.Fprintf(os.Stdout, "could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Finally, send a response to redirect the user to the "welcome" page
	// with the access token
	w.Header().Set("Location", "/welcome.html?access_token="+respBody.AccessToken)
	w.WriteHeader(http.StatusFound)
}
