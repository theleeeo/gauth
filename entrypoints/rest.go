package entrypoints

import (
	"net/http"

	"github.com/theleeeo/thor/app"
)

type restHandler struct {
	app        *app.App
	cookieName string
}

func NewRestHandler(app *app.App, cookieName string) *restHandler {
	return &restHandler{
		app:        app,
		cookieName: cookieName,
	}
}

func (h *restHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/public-key", h.PublicKey)
	mux.HandleFunc("/whoami", h.WhoAmI)
	mux.HandleFunc("/user/{id}", h.GetUserByID)
}

func (h *restHandler) PublicKey(w http.ResponseWriter, r *http.Request) {
	w.Write(h.app.PublicKey())
}

func (h *restHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(h.cookieName)
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
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	user, err := h.app.GetUserByID(r.Context(), id)
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err.Error() == "forbidden" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err.Error() == "unauthorized" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		respondError(w, err, http.StatusInternalServerError)
		return
	}

	respond(w, user)
}
