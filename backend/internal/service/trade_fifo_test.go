package service

import (
	"testing"
	"time"

	"proclients/backend/internal/model"
)

func TestComputeTradeProfitFIFO_UsesEarliestRemainingLot(t *testing.T) {
	product := "title:установка чехлов"
	sept1 := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	sept2 := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	sept3 := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	lots := []model.IncomingStockLot{
		{ProductKey: product, Quantity: 1, UnitCost: 500, InvoiceDate: sept1, CreatedAt: sept1},
		{ProductKey: product, Quantity: 1, UnitCost: 1000, InvoiceDate: sept2, CreatedAt: sept2},
	}
	sales := []model.OutgoingSaleLine{
		{
			InvoiceID:   "out-1",
			Title:       "Установка чехлов",
			ProductKey:  product,
			Quantity:    1,
			UnitPrice:   2000,
			InvoiceDate: sept3,
			CreatedAt:   sept3,
		},
	}

	from := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 3, 23, 59, 59, 0, time.UTC)
	totals, items := computeTradeProfitFIFO(lots, sales, from, to)

	if totals.InvoicesCount != 1 {
		t.Fatalf("invoicesCount = %d, want 1", totals.InvoicesCount)
	}
	if totals.Revenue != 2000 {
		t.Fatalf("revenue = %v, want 2000", totals.Revenue)
	}
	if totals.Cost != 500 {
		t.Fatalf("cost = %v, want 500", totals.Cost)
	}
	if totals.Profit != 1500 {
		t.Fatalf("profit = %v, want 1500", totals.Profit)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].CostPrice != 500 || items[0].SalePrice != 2000 || items[0].Profit != 1500 {
		t.Fatalf("item = %+v", items[0])
	}
}

func TestComputeTradeProfitFIFO_SplitsLotsWhenSaleConsumesBoth(t *testing.T) {
	product := "title:установка чехлов"
	sept1 := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	sept2 := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	sept3 := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	lots := []model.IncomingStockLot{
		{ProductKey: product, Quantity: 1, UnitCost: 500, InvoiceDate: sept1, CreatedAt: sept1},
		{ProductKey: product, Quantity: 1, UnitCost: 1000, InvoiceDate: sept2, CreatedAt: sept2},
	}
	sales := []model.OutgoingSaleLine{
		{
			InvoiceID:   "out-1",
			Title:       "Установка чехлов",
			ProductKey:  product,
			Quantity:    2,
			UnitPrice:   2000,
			InvoiceDate: sept3,
			CreatedAt:   sept3,
		},
	}

	from := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 3, 23, 59, 59, 0, time.UTC)
	totals, items := computeTradeProfitFIFO(lots, sales, from, to)

	if totals.Cost != 1500 {
		t.Fatalf("cost = %v, want 1500", totals.Cost)
	}
	if totals.Profit != 2500 {
		t.Fatalf("profit = %v, want 2500", totals.Profit)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2: %+v", len(items), items)
	}
}

func TestComputeTradeProfitFIFO_PreviousSaleConsumesEarliestLot(t *testing.T) {
	product := "title:установка чехлов"
	sept1 := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	sept2 := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	sept3 := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)

	lots := []model.IncomingStockLot{
		{ProductKey: product, Quantity: 1, UnitCost: 500, InvoiceDate: sept1, CreatedAt: sept1},
		{ProductKey: product, Quantity: 1, UnitCost: 1000, InvoiceDate: sept2, CreatedAt: sept2},
	}
	sales := []model.OutgoingSaleLine{
		{
			InvoiceID:   "out-old",
			Title:       "Установка чехлов",
			ProductKey:  product,
			Quantity:    1,
			UnitPrice:   1800,
			InvoiceDate: sept2,
			CreatedAt:   sept2.Add(time.Hour),
		},
		{
			InvoiceID:   "out-new",
			Title:       "Установка чехлов",
			ProductKey:  product,
			Quantity:    1,
			UnitPrice:   2000,
			InvoiceDate: sept3,
			CreatedAt:   sept3,
		},
	}

	from := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 3, 23, 59, 59, 0, time.UTC)
	totals, items := computeTradeProfitFIFO(lots, sales, from, to)

	if totals.Cost != 1000 {
		t.Fatalf("cost = %v, want 1000 from the later lot", totals.Cost)
	}
	if totals.Profit != 1000 {
		t.Fatalf("profit = %v, want 1000", totals.Profit)
	}
	if len(items) != 1 || items[0].CostPrice != 1000 {
		t.Fatalf("item = %+v", items)
	}
}
