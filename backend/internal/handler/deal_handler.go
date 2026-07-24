package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type DealHandler struct {
	service     *service.DealService
	attachments *service.AttachmentService
}

func NewDealHandler(service *service.DealService, attachments *service.AttachmentService) *DealHandler {
	return &DealHandler{service: service, attachments: attachments}
}

func (h *DealHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *DealHandler) CreateFromLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req model.CreateDealFromLeadInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if userID := auth.UserIDFromContext(r.Context()); userID != "" {
		req.CreatedBy = userID
	}
	item, err := h.service.CreateFromLead(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func (h *DealHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/deals/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		writeError(w, http.StatusBadRequest, "invalid deal id")
		return
	}
	dealID := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			if !requireAdmin(w, r) {
				return
			}
			if err := h.service.Delete(r.Context(), dealID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch parts[1] {
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
		item, err := h.service.UpdateStatus(r.Context(), dealID, req.ColumnID, failureReason)
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
			DealComments string `json:"dealComments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateComment(r.Context(), dealID, req.DealComments)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "profile":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			FirstName  string `json:"firstName"`
			Patronymic string `json:"patronymic"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateProfile(r.Context(), dealID, req.FirstName, req.Patronymic)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case "production-due-at":
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			DueAt *int64 `json:"dueAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.UpdateProductionDueAt(r.Context(), dealID, req.DueAt)
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
		item, err := h.service.UpdateProduction(r.Context(), dealID, req.Production)
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
		item, err := h.service.UpdatePickupDelivery(r.Context(), dealID, req.PickupDelivery)
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
		item, err := h.service.UpdateProducts(r.Context(), dealID, req.Products)
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
				return h.attachments.UploadToDeal(r.Context(), dealID, userID, filename, mimeType, content)
			},
			func() (any, error) {
				return h.attachments.ListByDeal(r.Context(), dealID)
			},
			func(userID string, count int) (any, error) {
				return h.attachments.LogDealUploadBatch(r.Context(), dealID, userID, count)
			},
		)
	case "activities":
		handleEntityActivities(
			w,
			r,
			func(userID string, text string) (any, error) {
				return h.service.AddComment(r.Context(), dealID, userID, text)
			},
			func() (any, error) {
				return h.service.ListActivities(r.Context(), dealID)
			},
		)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
