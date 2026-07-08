package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"proclients/backend/internal/service"
)

type EventsHandler struct {
	events *service.EventBus
}

func NewEventsHandler(events *service.EventBus) *EventsHandler {
	return &EventsHandler{events: events}
}

func (h *EventsHandler) LeadCreatedStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsubscribe := h.events.SubscribeLeadCreated()
	defer unsubscribe()

	// Initial hello (helps some proxies establish stream).
	fmt.Fprintf(w, "event: ready\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	keepAlive := time.NewTicker(25 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepAlive.C:
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		case event := <-ch:
			payload, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: lead-created\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

