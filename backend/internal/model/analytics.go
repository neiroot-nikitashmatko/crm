package model

import "time"

type TrafficSourceMetric struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

type LeadToDealConversion struct {
	LeadsCount     int     `json:"leadsCount"`
	ConvertedCount int     `json:"convertedCount"`
	Percent        float64 `json:"percent"`
}

type FailedLeadShare struct {
	LeadsCount  int     `json:"leadsCount"`
	FailedCount int     `json:"failedCount"`
	Percent     float64 `json:"percent"`
}

type FailedDealShare struct {
	DealsCount  int     `json:"dealsCount"`
	FailedCount int     `json:"failedCount"`
	Percent     float64 `json:"percent"`
}

type NomenclatureCount struct {
	Nomenclature string `json:"nomenclature"`
	Count        int    `json:"count"`
}

type ProductionCategoryMetric struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type EmployeeShareMetric struct {
	Employee string `json:"employee"`
	Count    int    `json:"count"`
}

type ClosedDealListItem struct {
	ID            string `json:"id"`
	DealNumber    int64  `json:"dealNumber"`
	FirstName     string `json:"firstName"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	Nomenclature  string `json:"nomenclature"`
	Category      string `json:"category"`
	Employee      string `json:"employee"`
	FailureReason string `json:"failureReason"`
	CreatedAt     int64  `json:"createdAt"`
	ClosedAt      int64  `json:"closedAt"`
}

type FailedLeadListItem struct {
	ID            string `json:"id"`
	LeadNumber    int64  `json:"leadNumber"`
	FirstName     string `json:"firstName"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	FailureReason string `json:"failureReason"`
	CreatedAt     int64  `json:"createdAt"`
}

type DealTrafficListItem struct {
	ID            string `json:"id"`
	DealNumber    int64  `json:"dealNumber"`
	FirstName     string `json:"firstName"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	TrafficSource string `json:"trafficSource"`
	CreatedAt     int64  `json:"createdAt"`
}

type TradeProfit struct {
	Profit        float64 `json:"profit"`
	Revenue       float64 `json:"revenue"`
	Cost          float64 `json:"cost"`
	InvoicesCount int     `json:"invoicesCount"`
}

type TradeProfitItem struct {
	ProductKey string  `json:"productKey"`
	Title      string  `json:"title"`
	Quantity   float64 `json:"quantity"`
	CostPrice  float64 `json:"costPrice"`
	SalePrice  float64 `json:"salePrice"`
	Profit     float64 `json:"profit"`
	HasCost    bool    `json:"hasCost"`
}

type IncomingStockLot struct {
	ProductKey    string
	Quantity      float64
	UnitCost      float64
	InvoiceDate   time.Time
	CreatedAt     time.Time
	InvoiceNumber int64
	Position      int
}

type OutgoingSaleLine struct {
	InvoiceID     string
	Title         string
	ProductKey    string
	Quantity      float64
	UnitPrice     float64
	InvoiceDate   time.Time
	CreatedAt     time.Time
	InvoiceNumber int64
	Position      int
}
