package entrypoints

import (
	"net/http"

	"github.com/theleeeo/thor/authorizer"
)

type restHandler struct {
	authorizer *authorizer.Authorizer
}

func NewRestHandler(a *authorizer.Authorizer) *restHandler {
	return &restHandler{
		authorizer: a,
	}
}

func (r *restHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /login", r.Login)
	mux.HandleFunc("GET /whoami", r.WhoAmI)
}

func (r *restHandler) Login(w http.ResponseWriter, req *http.Request) {
	loginReq, err := parse[LoginRequest](req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if loginReq.Username == "admin" && loginReq.Password == "admin" {
		token, err := r.authorizer.CreateToken(loginReq.Username, authorizer.RoleAdmin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, LoginResponse{
			Token: token,
		})
		return
	}

	if loginReq.Username == "user" && loginReq.Password == "user" {
		token, err := r.authorizer.CreateToken(loginReq.Username, authorizer.RoleUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, LoginResponse{
			Token: token,
		})
		return
	}

	http.Error(w, "invalid username or password", http.StatusUnauthorized)
}

func (r *restHandler) WhoAmI(w http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	token = token[len("Bearer "):]

	claims, err := r.authorizer.Decode(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	respond(w, WhoAmIResponse{
		Username: claims.UserID,
		Role:     string(claims.Role),
	})
}
