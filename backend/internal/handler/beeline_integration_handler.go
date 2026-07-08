package handler

import (
	"io"
	"net/http"
	"strings"

	"proclients/backend/internal/service"
)

type BeelineIntegrationHandler struct {
	service *service.BeelineIntegrationService
}

func NewBeelineIntegrationHandler(service *service.BeelineIntegrationService) *BeelineIntegrationHandler {
	return &BeelineIntegrationHandler{service: service}
}

func (h *BeelineIntegrationHandler) XSIEvents(w http.ResponseWriter, r *http.Request) {
	// Beeline may validate callback URL with GET/HEAD.
	// Always return 200 OK for validation requests.
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	secret := r.Header.Get("X-Beeline-Secret")
	if secret == "" {
		secret = r.URL.Query().Get("secret")
	}
	if secret == "" {
		// Support embedding secret in path: /api/v1/integrations/beeline/xsi-events/<secret>/...
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/integrations/beeline/xsi-events")
		path = strings.TrimPrefix(path, "/")
		if path != "" {
			parts := strings.SplitN(path, "/", 2)
			secret = parts[0]
		}
	}
	if !h.service.VerifySecret(secret) {
		// Do not reveal secret validity; do not block provider validation.
		// We simply ignore events without correct secret.
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "action": "ignored"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	result, err := h.service.HandleXSIEvent(r.Context(), body, r.Header.Get("Content-Type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

