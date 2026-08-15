package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
	"proclients/backend/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Collection(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.CreateUserInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Create(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	userID := strings.TrimSpace(parts[0])
	if len(parts) == 2 && parts[1] == "avatar" {
		h.avatar(w, r, userID)
		return
	}
	if len(parts) != 1 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if !requireAdmin(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := h.service.GetByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodPatch:
		var req model.UpdateUserInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Update(r.Context(), userID, req)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		err := h.service.Delete(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) avatar(w http.ResponseWriter, r *http.Request, userID string) {
	switch r.Method {
	case http.MethodGet:
		if strings.TrimSpace(auth.UserIDFromContext(r.Context())) == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}
		content, mimeType, err := h.service.GetAvatar(r.Context(), userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, repository.ErrUserAvatarNotFound) {
				writeError(w, http.StatusNotFound, "avatar not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.Header().Set("Cache-Control", "private, max-age=3600")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	case http.MethodPost:
		if !requireAdmin(w, r) {
			return
		}
		files, err := service.ParseMultipartFiles(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if len(files) == 0 {
			writeError(w, http.StatusBadRequest, "file is required")
			return
		}
		file := files[0]
		if err := h.service.SetAvatar(r.Context(), userID, file.Content, file.MimeType); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		item, err := h.service.GetByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if !requireAdmin(w, r) {
			return
		}
		if err := h.service.ClearAvatar(r.Context(), userID); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "Доступно только администратору")
		return false
	}
	return true
}
