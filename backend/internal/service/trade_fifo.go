package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"proclients/backend/internal/model"
)

const fifoQtyEpsilon = 1e-9

type fifoLot struct {
	remaining   float64
	unitCost    float64
	invoiceDate time.Time
	createdAt   time.Time
}

type fifoAlloc struct {
	productKey string
	title      string
	quantity   float64
	unitCost   float64
	hasCost    bool
	salePrice  float64
}

type fifoGroupKey struct {
	productKey string
	hasCost    bool
	costCents  int64
}

type fifoGroup struct {
	title    string
	titleQty float64
	quantity float64
	revenue  float64
	cost     float64
}

func computeTradeProfitFIFO(
	lots []model.IncomingStockLot,
	sales []model.OutgoingSaleLine,
	from time.Time,
	to time.Time,
) (model.TradeProfit, []model.TradeProfitItem) {
	remainingByProduct := make(map[string][]*fifoLot)
	for _, lot := range lots {
		if lot.Quantity <= fifoQtyEpsilon {
			continue
		}
		remainingByProduct[lot.ProductKey] = append(remainingByProduct[lot.ProductKey], &fifoLot{
			remaining:   lot.Quantity,
			unitCost:    lot.UnitCost,
			invoiceDate: lot.InvoiceDate,
			createdAt:   lot.CreatedAt,
		})
	}
	for _, productLots := range remainingByProduct {
		sort.SliceStable(productLots, func(i int, j int) bool {
			return incomingBefore(productLots[i].invoiceDate, productLots[i].createdAt, productLots[j].invoiceDate, productLots[j].createdAt)
		})
	}

	orderedSales := append([]model.OutgoingSaleLine(nil), sales...)
	sort.SliceStable(orderedSales, func(i int, j int) bool {
		left, right := orderedSales[i], orderedSales[j]
		if !left.InvoiceDate.Equal(right.InvoiceDate) {
			return left.InvoiceDate.Before(right.InvoiceDate)
		}
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.Before(right.CreatedAt)
		}
		if left.InvoiceNumber != right.InvoiceNumber {
			return left.InvoiceNumber < right.InvoiceNumber
		}
		return left.Position < right.Position
	})

	periodAllocs := make([]fifoAlloc, 0)
	invoiceIDs := make(map[string]struct{})

	for _, sale := range orderedSales {
		inPeriod := !sale.InvoiceDate.Before(from) && !sale.InvoiceDate.After(to)
		if inPeriod && sale.InvoiceID != "" {
			invoiceIDs[sale.InvoiceID] = struct{}{}
		}

		qtyLeft := sale.Quantity
		for _, lot := range remainingByProduct[sale.ProductKey] {
			if qtyLeft <= fifoQtyEpsilon {
				break
			}
			if !lotAvailableForSale(lot, sale) {
				break
			}
			if lot.remaining <= fifoQtyEpsilon {
				continue
			}

			take := math.Min(lot.remaining, qtyLeft)
			lot.remaining -= take
			qtyLeft -= take
			if !inPeriod {
				continue
			}
			periodAllocs = append(periodAllocs, fifoAlloc{
				productKey: sale.ProductKey,
				title:      sale.Title,
				quantity:   take,
				unitCost:   lot.unitCost,
				hasCost:    true,
				salePrice:  sale.UnitPrice,
			})
		}

		if inPeriod && qtyLeft > fifoQtyEpsilon {
			periodAllocs = append(periodAllocs, fifoAlloc{
				productKey: sale.ProductKey,
				title:      sale.Title,
				quantity:   qtyLeft,
				unitCost:   0,
				hasCost:    false,
				salePrice:  sale.UnitPrice,
			})
		}
	}

	return aggregateFIFOPeriod(periodAllocs, len(invoiceIDs))
}

func incomingBefore(leftDate time.Time, leftCreated time.Time, rightDate time.Time, rightCreated time.Time) bool {
	if !leftDate.Equal(rightDate) {
		return leftDate.Before(rightDate)
	}
	return leftCreated.Before(rightCreated)
}

func lotAvailableForSale(lot *fifoLot, sale model.OutgoingSaleLine) bool {
	if lot.invoiceDate.Before(sale.InvoiceDate) {
		return true
	}
	if lot.invoiceDate.After(sale.InvoiceDate) {
		return false
	}
	return !lot.createdAt.After(sale.CreatedAt)
}

func aggregateFIFOPeriod(allocs []fifoAlloc, invoicesCount int) (model.TradeProfit, []model.TradeProfitItem) {
	groups := make(map[fifoGroupKey]*fifoGroup)
	order := make([]fifoGroupKey, 0)
	totals := model.TradeProfit{InvoicesCount: invoicesCount}

	for _, alloc := range allocs {
		totals.Revenue += alloc.quantity * alloc.salePrice
		if alloc.hasCost {
			totals.Cost += alloc.quantity * alloc.unitCost
		}

		key := fifoGroupKey{
			productKey: alloc.productKey,
			hasCost:    alloc.hasCost,
			costCents:  moneyCents(alloc.unitCost),
		}
		group, ok := groups[key]
		if !ok {
			group = &fifoGroup{}
			groups[key] = group
			order = append(order, key)
		}
		group.quantity += alloc.quantity
		group.revenue += alloc.quantity * alloc.salePrice
		if alloc.hasCost {
			group.cost += alloc.quantity * alloc.unitCost
		}
		title := strings.TrimSpace(alloc.title)
		if title != "" && alloc.quantity >= group.titleQty {
			group.title = title
			group.titleQty = alloc.quantity
		}
	}

	totals.Profit = totals.Revenue - totals.Cost

	items := make([]model.TradeProfitItem, 0, len(order))
	for _, key := range order {
		group := groups[key]
		title := strings.TrimSpace(group.title)
		if title == "" {
			title = "Без названия"
		}
		salePrice := 0.0
		if group.quantity > fifoQtyEpsilon {
			salePrice = group.revenue / group.quantity
		}
		costPrice := 0.0
		if key.hasCost && group.quantity > fifoQtyEpsilon {
			costPrice = group.cost / group.quantity
		}
		items = append(items, model.TradeProfitItem{
			ProductKey: key.productKey,
			Title:      title,
			Quantity:   group.quantity,
			CostPrice:  costPrice,
			SalePrice:  salePrice,
			Profit:     group.revenue - group.cost,
			HasCost:    key.hasCost,
		})
	}

	sort.SliceStable(items, func(i int, j int) bool {
		if items[i].Profit != items[j].Profit {
			return items[i].Profit > items[j].Profit
		}
		if items[i].Title != items[j].Title {
			return items[i].Title < items[j].Title
		}
		return items[i].CostPrice < items[j].CostPrice
	})

	return totals, items
}

func moneyCents(value float64) int64 {
	return int64(math.Round(value * 100))
}
