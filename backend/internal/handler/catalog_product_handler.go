package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type CatalogProductHandler struct {
	service *service.CatalogProductService
}

func NewCatalogProductHandler(service *service.CatalogProductService) *CatalogProductHandler {
	return &CatalogProductHandler{service: service}
}

func (h *CatalogProductHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.UpsertCatalogProductInput
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

func (h *CatalogProductHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/catalog-products/")
	productID := strings.TrimSpace(path)
	if productID == "" || strings.Contains(productID, "/") {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req model.UpsertCatalogProductInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Update(r.Context(), productID, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if err := h.service.Delete(r.Context(), productID); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
