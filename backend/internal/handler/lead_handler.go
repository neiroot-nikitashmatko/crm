package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type LeadHandler struct {
	service *service.LeadService
}

func NewLeadHandler(service *service.LeadService) *LeadHandler {
	return &LeadHandler{service: service}
}

func (h *LeadHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.CreateLeadInput
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

func (h *LeadHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid lead id")
		return
	}
	leadID := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			if err := h.service.Delete(r.Context(), leadID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	action := parts[1]
	switch action {
	case "status":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ColumnID      string `json:"columnId"`
			FailureReason string `json:"failureReason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		var failureReason *string
		if req.ColumnID == "failed" {
			trimmed := strings.TrimSpace(req.FailureReason)
			if trimmed == "" {
				writeError(w, http.StatusBadRequest, "failureReason is required")
				return
			}
			failureReason = &trimmed
		}
		item, err := h.service.UpdateColumn(r.Context(), leadID, req.ColumnID, failureReason)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "comment":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			LeadComments string `json:"leadComments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateComment(r.Context(), leadID, req.LeadComments)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "pickup-delivery":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			PickupDelivery model.PickupDelivery `json:"pickupDelivery"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdatePickupDelivery(r.Context(), leadID, req.PickupDelivery)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "products":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Products []model.DealProduct `json:"products"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateProducts(r.Context(), leadID, req.Products)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "production":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Production model.DealProduction `json:"production"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateProduction(r.Context(), leadID, req.Production)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
