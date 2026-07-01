package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type TaskHandler struct {
	service     *service.TaskService
	attachments *service.AttachmentService
}

func NewTaskHandler(service *service.TaskService, attachments *service.AttachmentService) *TaskHandler {
	return &TaskHandler{service: service, attachments: attachments}
}

func (h *TaskHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.CreateTaskInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if userID := auth.UserIDFromContext(r.Context()); userID != "" {
			req.CreatedBy = userID
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

func (h *TaskHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	taskID := parts[0]

	if len(parts) == 1 {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		input := model.UpdateTaskInput{}
		if value, ok := raw["title"].(string); ok {
			input.Title = &value
		}
		if value, ok := raw["text"].(string); ok {
			input.Text = &value
		}
		if value, ok := raw["dueAt"]; ok {
			input.HasDueAt = true
			if value == nil {
				input.DueAt = nil
			} else if numeric, ok := value.(float64); ok {
				parsed := int64(numeric)
				input.DueAt = &parsed
			}
		}
		item, err := h.service.Update(r.Context(), taskID, input)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
		return
	}

	switch parts[1] {
	case "complete":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		item, err := h.service.Complete(r.Context(), taskID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "attachments":
		handleEntityAttachments(
			w,
			r,
			func(userID string, filename string, mimeType string, content []byte) (any, error) {
				return h.attachments.UploadToTask(r.Context(), taskID, userID, filename, mimeType, content)
			},
			func() (any, error) {
				return h.attachments.ListByTask(r.Context(), taskID)
			},
			func(userID string, count int) (any, error) {
				return h.attachments.LogTaskUploadBatch(r.Context(), taskID, userID, count)
			},
		)
	case "activities":
		handleEntityActivities(
			w,
			r,
			func(userID string, text string) (any, error) {
				return h.service.AddComment(r.Context(), taskID, userID, text)
			},
			func() (any, error) {
				return h.service.ListActivities(r.Context(), taskID)
			},
		)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *TaskHandler) CompleteByLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		LeadID string `json:"leadId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.CompleteByLead(r.Context(), req.LeadID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
