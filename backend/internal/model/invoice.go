package model

type InvoiceItem struct {
	CatalogProductID string  `json:"catalogProductId,omitempty"`
	Title            string  `json:"title"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unitPrice"`
}

type IncomingInvoice struct {
	ID            string        `json:"id"`
	InvoiceNumber int64         `json:"invoiceNumber"`
	Date          int64         `json:"date"`
	SupplierID    string        `json:"supplierId"`
	Items         []InvoiceItem `json:"items"`
	Total         float64       `json:"total"`
	Comment       string        `json:"comment"`
	CreatedAt     int64         `json:"createdAt"`
}

type UpsertIncomingInvoiceInput struct {
	Date       int64         `json:"date"`
	SupplierID string        `json:"supplierId"`
	Items      []InvoiceItem `json:"items"`
	Comment    string        `json:"comment"`
}

type OutgoingInvoice struct {
	ID            string        `json:"id"`
	InvoiceNumber int64         `json:"invoiceNumber"`
	Date          int64         `json:"date"`
	DealID        string        `json:"dealId"`
	Items         []InvoiceItem `json:"items"`
	Total         float64       `json:"total"`
	Comment       string        `json:"comment"`
	CreatedAt     int64         `json:"createdAt"`
}

type UpsertOutgoingInvoiceInput struct {
	Date    int64         `json:"date"`
	DealID  string        `json:"dealId"`
	Items   []InvoiceItem `json:"items"`
	Comment string        `json:"comment"`
}
