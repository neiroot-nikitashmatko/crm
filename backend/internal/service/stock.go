package service

import (
	"strings"

	"proclients/backend/internal/model"
)

func stockMovementKey(catalogProductID string, title string) string {
	id := strings.TrimSpace(catalogProductID)
	if id != "" {
		return "id:" + id
	}
	return "title:" + strings.ToLower(strings.TrimSpace(title))
}

func applyStockMovement(quantities map[string]float64, item model.InvoiceItem, sign float64) {
	title := strings.TrimSpace(item.Title)
	if title == "" && strings.TrimSpace(item.CatalogProductID) == "" {
		return
	}
	key := stockMovementKey(item.CatalogProductID, title)
	quantities[key] += sign * item.Quantity
}

func buildStockQuantities(
	incoming []model.IncomingInvoice,
	outgoing []model.OutgoingInvoice,
	excludeOutgoingID string,
) map[string]float64 {
	quantities := make(map[string]float64)

	for _, invoice := range incoming {
		for _, item := range invoice.Items {
			applyStockMovement(quantities, item, 1)
		}
	}

	excludeOutgoingID = strings.TrimSpace(excludeOutgoingID)
	for _, invoice := range outgoing {
		if excludeOutgoingID != "" && invoice.ID == excludeOutgoingID {
			continue
		}
		for _, item := range invoice.Items {
			applyStockMovement(quantities, item, -1)
		}
	}

	return quantities
}

func outgoingStockShortageTitles(
	items []model.InvoiceItem,
	quantities map[string]float64,
) []string {
	requested := make(map[string]float64)
	titles := make(map[string]string)

	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" && strings.TrimSpace(item.CatalogProductID) == "" {
			continue
		}
		key := stockMovementKey(item.CatalogProductID, title)
		requested[key] += item.Quantity
		if title != "" {
			titles[key] = title
		} else if titles[key] == "" {
			titles[key] = "Товар"
		}
	}

	shortages := make([]string, 0)
	seen := make(map[string]struct{})
	for key, quantity := range requested {
		if quantity <= quantities[key] {
			continue
		}
		title := titles[key]
		if _, exists := seen[title]; exists {
			continue
		}
		seen[title] = struct{}{}
		shortages = append(shortages, title)
	}
	return shortages
}
