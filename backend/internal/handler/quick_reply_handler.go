package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type QuickReplyHandler struct {
	service *service.QuickReplyService
}

func NewQuickReplyHandler(service *service.QuickReplyService) *QuickReplyHandler {
	return &QuickReplyHandler{service: service}
}

func (h *QuickReplyHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.CreateQuickReplySectionInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.CreateSection(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *QuickReplyHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/quick-reply-sections/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid section id")
		return
	}
	sectionID := strings.TrimSpace(parts[0])

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodPatch:
			var req model.UpdateQuickReplySectionInput
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json")
				return
			}
			item, err := h.service.UpdateSection(r.Context(), sectionID, req)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"item": item})
		case http.MethodDelete:
			if err := h.service.DeleteSection(r.Context(), sectionID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 2 && parts[1] == "replies" && r.Method == http.MethodPost {
		var req model.CreateQuickReplyInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.CreateReply(r.Context(), sectionID, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *QuickReplyHandler) ReplyItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/quick-replies/")
	replyID := strings.TrimSpace(path)
	if replyID == "" || strings.Contains(replyID, "/") {
		writeError(w, http.StatusBadRequest, "invalid reply id")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req model.UpdateQuickReplyInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateReply(r.Context(), replyID, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if err := h.service.DeleteReply(r.Context(), replyID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
