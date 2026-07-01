package handler

import (
	"encoding/json"
	"net/http"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/repository"
	"proclients/backend/internal/service"
)

type AuthHandler struct {
	service    *service.AuthService
	jwtManager *auth.Manager
}

func NewAuthHandler(service *service.AuthService, jwtManager *auth.Manager) *AuthHandler {
	return &AuthHandler{
		service:    service,
		jwtManager: jwtManager,
	}
}

type loginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.service.Login(r.Context(), req.Phone, req.Password)
	if err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.jwtManager.Issue(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}
