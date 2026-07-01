package model

type CatalogProduct struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	Category  string  `json:"category"`
	Cost      float64 `json:"cost"`
	CreatedAt int64   `json:"createdAt"`
	UpdatedAt int64   `json:"updatedAt"`
}

type UpsertCatalogProductInput struct {
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Category string  `json:"category"`
	Cost     float64 `json:"cost"`
}
