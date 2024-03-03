package entrypoints

import (
	"net/http"

	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/sdk"
)

type restHandler struct {
	app *app.App
}

func NewRestHandler(app *app.App) *restHandler {
	return &restHandler{
		app: app,
	}
}

func (h *restHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /public-key", h.PublicKey)

	mux.HandleFunc("GET /whoami", h.WhoAmI)
	mux.HandleFunc("GET /user/", h.GetUserByID)
}

func (h *restHandler) PublicKey(w http.ResponseWriter, r *http.Request) {
	w.Write(h.app.PublicKey())
}

func (h *restHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("thor_token")
	if err != nil {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	user, err := h.app.WhoAmI(r.Context(), token.Value)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError)
		return
	}

	respond(w, user)
}

func (h *restHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	//The path will be /user/{id}
	id := r.URL.Path[len("/user/"):]
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	claims, err := sdk.ExtractClaims(r, h.app.PublicKey())
	if err != nil {
		respondError(w, err, http.StatusUnauthorized)
		return
	}

	if claims.UserID != id {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	user, err := h.app.GetUserByID(r.Context(), id)
	if err != nil {
		respondError(w, err, http.StatusInternalServerError)
		return
	}

	respond(w, user)
}
