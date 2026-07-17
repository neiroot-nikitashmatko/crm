package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type SalaryEntryHandler struct {
	service *service.SalaryEntryService
}

func NewSalaryEntryHandler(service *service.SalaryEntryService) *SalaryEntryHandler {
	return &SalaryEntryHandler{service: service}
}

func (h *SalaryEntryHandler) Collection(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	role := auth.RoleFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "authorization required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context(), userID, role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.UpsertSalaryEntryInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Create(r.Context(), userID, role, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *SalaryEntryHandler) Item(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	role := auth.RoleFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "authorization required")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/salary-entries/")
	entryID := strings.TrimSpace(path)
	if entryID == "" || strings.Contains(entryID, "/") {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req model.UpsertSalaryEntryInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Update(r.Context(), entryID, userID, role, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if err := h.service.Delete(r.Context(), entryID, userID, role); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
