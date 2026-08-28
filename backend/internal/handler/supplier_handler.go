package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
	"proclients/backend/internal/service"
)

type SupplierHandler struct {
	service *service.SupplierService
}

func NewSupplierHandler(service *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: service}
}

func (h *SupplierHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.UpsertSupplierInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
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

func (h *SupplierHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/suppliers/")
	supplierID := strings.TrimSpace(path)
	if supplierID == "" || strings.Contains(supplierID, "/") {
		writeError(w, http.StatusBadRequest, "некорректный идентификатор поставщика")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req model.UpsertSupplierInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Update(r.Context(), supplierID, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if err := h.service.Delete(r.Context(), supplierID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeTradeWriteError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrSupplierNotFound) ||
		errors.Is(err, repository.ErrIncomingInvoiceNotFound) ||
		errors.Is(err, repository.ErrOutgoingInvoiceNotFound) ||
		errors.Is(err, repository.ErrOutgoingInvoiceDealTaken) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}
