package handler

import (
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
)

func NewRouter(
	authHandler *AuthHandler,
	leadHandler *LeadHandler,
	dealHandler *DealHandler,
	taskHandler *TaskHandler,
	catalogProductHandler *CatalogProductHandler,
	userHandler *UserHandler,
	attachmentHandler *AttachmentHandler,
	beelineHandler *BeelineIntegrationHandler,
	eventsHandler *EventsHandler,
	jwtManager *auth.Manager,
	corsOrigins []string,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)

	mux.HandleFunc("/api/v1/leads", leadHandler.Collection)
	mux.HandleFunc("/api/v1/leads/", leadHandler.Item)

	mux.HandleFunc("/api/v1/deals", dealHandler.Collection)
	mux.HandleFunc("/api/v1/deals/from-lead", dealHandler.CreateFromLead)
	mux.HandleFunc("/api/v1/deals/", dealHandler.Item)

	mux.HandleFunc("/api/v1/tasks", taskHandler.Collection)
	mux.HandleFunc("/api/v1/tasks/complete-by-lead", taskHandler.CompleteByLead)
	mux.HandleFunc("/api/v1/tasks/", taskHandler.Item)

	mux.HandleFunc("/api/v1/catalog-products", catalogProductHandler.Collection)
	mux.HandleFunc("/api/v1/catalog-products/", catalogProductHandler.Item)

	mux.HandleFunc("/api/v1/users", userHandler.Collection)
	mux.HandleFunc("/api/v1/users/", userHandler.Item)

	mux.HandleFunc("/api/v1/attachments/", attachmentHandler.Item)

	mux.HandleFunc("/api/v1/integrations/beeline/xsi-events", beelineHandler.XSIEvents)
	// Allow Beeline to append extra path segments (e.g. /null) and allow embedding secret in path.
	mux.HandleFunc("/api/v1/integrations/beeline/xsi-events/", beelineHandler.XSIEvents)

	// Authenticated SSE stream with internal events.
	mux.HandleFunc("/api/v1/events/leads", eventsHandler.LeadCreatedStream)

	return withCORS(withAuth(jwtManager, mux), corsOrigins)
}

func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	normalized := make([]string, 0, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			for _, allowed := range normalized {
				if allowed == "*" || allowed == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
