package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/avito"
	"proclients/backend/internal/service"
)

type AvitoIntegrationHandler struct {
	service       *service.AvitoIntegrationService
	notifications *service.NotificationService
}

func NewAvitoIntegrationHandler(
	service *service.AvitoIntegrationService,
	notifications *service.NotificationService,
) *AvitoIntegrationHandler {
	return &AvitoIntegrationHandler{service: service, notifications: notifications}
}

func (h *AvitoIntegrationHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	// Avito / tunnels may probe with GET/HEAD.
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	secret := r.URL.Query().Get("secret")
	if secret == "" {
		secret = r.URL.Query().Get("token")
	}
	if secret == "" {
		secret = r.Header.Get("X-Avito-Secret")
	}
	if !h.service.VerifySecret(secret) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "action": "ignored"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	result, err := h.service.HandleWebhook(r.Context(), body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type avitoSubscribeRequest struct {
	URL string `json:"url"`
}

func (h *AvitoIntegrationHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !h.service.Enabled() {
		writeError(w, http.StatusBadRequest, "avito integration is not configured")
		return
	}

	var payload avitoSubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	url := strings.TrimSpace(payload.URL)
	if url == "" {
		url = strings.TrimSpace(r.URL.Query().Get("url"))
	}
	if url == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	if err := h.service.SubscribeWebhook(r.Context(), url); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "url": url})
}

func (h *AvitoIntegrationHandler) ChatsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	items, err := h.service.ListChats(r.Context(), auth.UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *AvitoIntegrationHandler) LeadChat(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/integrations/avito/chats/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "leadId is required")
		return
	}
	leadID := strings.TrimSpace(parts[0])

	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		bundle, err := h.service.GetLeadChatBundle(r.Context(), leadID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, bundle)
		return
	}

	if len(parts) == 2 && parts[1] == "read" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if h.notifications == nil {
			writeError(w, http.StatusInternalServerError, "notifications service is not configured")
			return
		}
		if err := h.notifications.MarkAvitoChatRead(r.Context(), auth.UserIDFromContext(r.Context()), leadID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	if len(parts) == 2 && parts[1] == "messages" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		text := ""
		var files []avito.UploadFile
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(25 << 20); err != nil {
				writeError(w, http.StatusBadRequest, "invalid multipart body")
				return
			}
			text = strings.TrimSpace(r.FormValue("text"))
			if r.MultipartForm != nil {
				for _, headers := range r.MultipartForm.File {
					for _, header := range headers {
						file, openErr := header.Open()
						if openErr != nil {
							writeError(w, http.StatusBadRequest, "failed to read file")
							return
						}
						data, readErr := io.ReadAll(file)
						_ = file.Close()
						if readErr != nil {
							writeError(w, http.StatusBadRequest, "failed to read file")
							return
						}
						files = append(files, avito.UploadFile{
							Filename:    header.Filename,
							ContentType: header.Header.Get("Content-Type"),
							Data:        data,
						})
					}
				}
			}
		} else {
			var payload struct {
				Text string `json:"text"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json body")
				return
			}
			text = payload.Text
		}

		items, err := h.service.SendMessageToLead(r.Context(), leadID, text, files)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
