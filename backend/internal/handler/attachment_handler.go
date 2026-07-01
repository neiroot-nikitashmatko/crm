package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/repository"
	"proclients/backend/internal/service"
)

type AttachmentHandler struct {
	service *service.AttachmentService
}

func NewAttachmentHandler(service *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{service: service}
}

func (h *AttachmentHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/attachments/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	attachmentID := parts[0]
	if len(parts) == 2 && parts[1] == "content" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		item, err := h.service.GetContent(r.Context(), attachmentID)
		if err != nil {
			if errors.Is(err, repository.ErrAttachmentNotFound) {
				writeError(w, http.StatusNotFound, "attachment not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.Header().Set("Content-Type", item.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", item.Name))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(item.Content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(item.Content)
		return
	}

	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	activity, err := h.service.Delete(r.Context(), attachmentID, auth.UserIDFromContext(r.Context()))
	if err != nil {
		if errors.Is(err, repository.ErrAttachmentNotFound) {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload := map[string]any{"ok": true}
	if activity != nil {
		payload["activity"] = activity
	}
	writeJSON(w, http.StatusOK, payload)
}

func handleEntityAttachments(
	w http.ResponseWriter,
	r *http.Request,
	upload func(userID string, filename string, mimeType string, content []byte) (any, error),
	list func() (any, error),
	afterBatchUpload func(userID string, count int) (any, error),
) {
	switch r.Method {
	case http.MethodGet:
		items, err := list()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		userID := auth.UserIDFromContext(r.Context())
		if strings.TrimSpace(userID) == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}
		files, err := service.ParseMultipartFiles(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		items := make([]any, 0, len(files))
		for _, file := range files {
			item, err := upload(userID, file.Filename, file.MimeType, file.Content)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			items = append(items, item)
		}
		payload := map[string]any{"items": items}
		if afterBatchUpload != nil && len(files) > 0 {
			activity, err := afterBatchUpload(userID, len(files))
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if activity != nil {
				payload["activity"] = activity
			}
		}
		writeJSON(w, http.StatusOK, payload)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleEntityActivities(
	w http.ResponseWriter,
	r *http.Request,
	createComment func(userID string, text string) (any, error),
	list func() (any, error),
) {
	switch r.Method {
	case http.MethodGet:
		items, err := list()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		userID := auth.UserIDFromContext(r.Context())
		if strings.TrimSpace(userID) == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := createComment(userID, req.Text)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
