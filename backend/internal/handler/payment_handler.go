package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"proclients/backend/internal/auth"
	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) Collection(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	role := auth.RoleFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "authorization required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.service.List(r.Context(), userID, role)
		if err != nil {
			writePaymentError(w, err, true)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req model.CreatePaymentInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		item, err := h.service.Create(r.Context(), userID, role, req)
		if err != nil {
			writePaymentError(w, err, false)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *PaymentHandler) Item(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	role := auth.RoleFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "authorization required")
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/payments/"), "/")
	parts := strings.Split(path, "/")
	paymentID := strings.TrimSpace(parts[0])
	if paymentID == "" || len(parts) != 1 {
		writeError(w, http.StatusBadRequest, "некорректный идентификатор оплаты")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req model.PatchPaymentInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if req.IsClosed != nil && strings.TrimSpace(req.ShortTitle) == "" && req.Date == 0 {
			item, err := h.service.SetClosed(r.Context(), paymentID, userID, role, *req.IsClosed)
			if err != nil {
				writePaymentError(w, err, false)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"item": item})
			return
		}

		item, err := h.service.Update(r.Context(), paymentID, userID, role, model.CreatePaymentInput{
			Date:         req.Date,
			RemindAt:     req.RemindAt,
			PayerID:      req.PayerID,
			Counterparty: req.Counterparty,
			Amount:       req.Amount,
			ShortTitle:   req.ShortTitle,
			Comment:      req.Comment,
		})
		if err != nil {
			writePaymentError(w, err, false)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	case http.MethodDelete:
		if err := h.service.Delete(r.Context(), paymentID, userID, role); err != nil {
			writePaymentError(w, err, false)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writePaymentError(w http.ResponseWriter, err error, isList bool) {
	switch {
	case errors.Is(err, service.ErrPaymentForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrPaymentNotFound):
		writeError(w, http.StatusBadRequest, err.Error())
	case err != nil && err.Error() == "authorization required":
		writeError(w, http.StatusUnauthorized, err.Error())
	case isList:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
