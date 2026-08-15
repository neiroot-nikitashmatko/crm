package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"proclients/backend/internal/model"
	"proclients/backend/internal/service"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) LeadsTraffic(w http.ResponseWriter, r *http.Request) {
	h.traffic(w, r, h.service.LeadsTraffic)
}

func (h *AnalyticsHandler) DealsTraffic(w http.ResponseWriter, r *http.Request) {
	h.traffic(w, r, h.service.DealsTraffic)
}

func (h *AnalyticsHandler) LeadToDealConversion(w http.ResponseWriter, r *http.Request) {
	h.share(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.LeadToDealConversion(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) FailedLeadShare(w http.ResponseWriter, r *http.Request) {
	h.share(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.FailedLeadShare(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) FailedDealShare(w http.ResponseWriter, r *http.Request) {
	h.share(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.FailedDealShare(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) ClosedDealsProductionShare(w http.ResponseWriter, r *http.Request) {
	h.items(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.ClosedDealsProductionShare(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) ClosedDealsEmployeeShare(w http.ResponseWriter, r *http.Request) {
	h.items(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.ClosedDealsEmployeeShare(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) ClosedDealsList(w http.ResponseWriter, r *http.Request) {
	requireEmployee := r.URL.Query().Get("requireEmployee") == "1"
	requireProduction := r.URL.Query().Get("requireProduction") == "1"
	h.items(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return h.service.ClosedDealsList(ctx, fromMs, toMs, requireEmployee, requireProduction)
	})
}

func (h *AnalyticsHandler) share(
	w http.ResponseWriter,
	r *http.Request,
	load func(context.Context, int64, int64) (any, error),
) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	fromMs, toMs, ok := parseAnalyticsPeriodQuery(w, r)
	if !ok {
		return
	}

	item, err := load(r.Context(), fromMs, toMs)
	if err != nil {
		if errors.Is(err, service.ErrAnalyticsPeriodRequired) || errors.Is(err, service.ErrInvalidAnalyticsPeriod) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func parseAnalyticsPeriodQuery(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	fromMs, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "укажите период")
		return 0, 0, false
	}
	toMs, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "укажите период")
		return 0, 0, false
	}
	return fromMs, toMs, true
}

func (h *AnalyticsHandler) traffic(
	w http.ResponseWriter,
	r *http.Request,
	load func(context.Context, int64, int64) ([]model.TrafficSourceMetric, error),
) {
	h.items(w, r, func(ctx context.Context, fromMs int64, toMs int64) (any, error) {
		return load(ctx, fromMs, toMs)
	})
}

func (h *AnalyticsHandler) items(
	w http.ResponseWriter,
	r *http.Request,
	load func(context.Context, int64, int64) (any, error),
) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	fromMs, toMs, ok := parseAnalyticsPeriodQuery(w, r)
	if !ok {
		return
	}

	items, err := load(r.Context(), fromMs, toMs)
	if err != nil {
		if errors.Is(err, service.ErrAnalyticsPeriodRequired) || errors.Is(err, service.ErrInvalidAnalyticsPeriod) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
