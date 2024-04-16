package entrypoints

import (
	"log/slog"
	"net/http"

	"github.com/theleeeo/thor/app"
	"github.com/theleeeo/thor/models"
	"github.com/theleeeo/thor/repo"
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
	mux.HandleFunc("GET /public-key", h.PublicKey)
	mux.HandleFunc("GET /whoami", h.WhoAmI)

	mux.HandleFunc("GET /users/{id}", h.GetUserByID)
	mux.HandleFunc("GET /users/{id}/permissions", h.GetPermissionsOfUser)
	mux.HandleFunc("GET /users", h.ListUsers)
	mux.HandleFunc("PATCH /users/{id}/roles/{role_id}", h.AssignRole)
	mux.HandleFunc("DELETE /users/{id}/roles/{role_id}", h.RemoveRole)
	mux.HandleFunc("GET /users/{id}/roles", h.GetRolesOfUser)

	mux.HandleFunc("GET /roles", h.ListRoles)
	mux.HandleFunc("GET /roles/{id}", h.GetRoleByID)
	mux.HandleFunc("POST /roles", h.CreateRole)
	mux.HandleFunc("GET /roles/{id}/permissions", h.GetPermissionsOfRole)
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

func (h *restHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	role, err := h.app.GetRoleByID(r.Context(), id)
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

	respond(w, role)
}

func (h *restHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.app.ListUsers(r.Context(), repo.ListUsersParams{})
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

	respond(w, users)
}

func (h *restHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.app.ListRoles(r.Context(), repo.ListRolesParams{})
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

	respond(w, roles)
}

type CreateRoleParams struct {
	Name        string            `json:"name"`
	Permissions map[string]string `json:"permissions"`
}

func (h *restHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	createRoleParams, err := parse[CreateRoleParams](r)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var permissions []models.Permission
	for k, v := range createRoleParams.Permissions {
		permissions = append(permissions, models.Permission{Key: k, Val: v})
	}

	role, err := h.app.CreateRole(r.Context(), models.Role{Name: createRoleParams.Name}, permissions)
	if err != nil {
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

	respond(w, role)
}

func (h *restHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	role_id := r.PathValue("role_id")
	if role_id == "" {
		http.Error(w, "missing role_id", http.StatusBadRequest)
		return
	}

	err := h.app.AssignRole(r.Context(), id, role_id)
	if err != nil {
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

	respond(w, nil)
}

func (h *restHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	role_id := r.PathValue("role_id")
	if role_id == "" {
		http.Error(w, "missing role_id", http.StatusBadRequest)
		return
	}

	err := h.app.RemoveRole(r.Context(), id, role_id)
	if err != nil {
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

	respond(w, nil)
}

func (h *restHandler) GetRolesOfUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	roles, err := h.app.GetRolesOfUser(r.Context(), id)
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

	respond(w, roles)
}

func (h *restHandler) GetPermissionsOfRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	permissions, err := h.app.GetPermissionsOfRole(r.Context(), id)
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

	respond(w, permissions)
}

func (h *restHandler) GetPermissionsOfUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	permissions, err := h.app.GetPermissionsOfUser(r.Context(), id)
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

	respond(w, permissions)
}
