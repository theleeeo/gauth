package entrypoints

import (
	"net/http"

	"github.com/theleeeo/thor/app"
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
	mux.HandleFunc("GET /whoami", h.WhoAmI)
	mux.HandleFunc("GET /user", h.GetUserByProviderID)
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

type IdReq struct {
	ID string `json:"id"`
}

func (h *restHandler) GetUserByProviderID(w http.ResponseWriter, r *http.Request) {
	user, err := h.app.GetUserByProviderID(r.Context(), "5105280")
	if err != nil {
		respondError(w, err, http.StatusInternalServerError)
		return
	}

	respond(w, user)
}
