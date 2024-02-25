package entrypoints

import (
	"encoding/json"
	"net/http"
)

func parse[T any](r *http.Request) (v *T, err error) {
	err = json.NewDecoder(r.Body).Decode(&v)
	return v, err
}

func respond(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type WhoAmIResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}
