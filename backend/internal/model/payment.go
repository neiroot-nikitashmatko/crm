package model

type Payment struct {
	ID           string  `json:"id"`
	Date         int64   `json:"date"`
	RemindAt     *int64  `json:"remindAt"`
	PayerID      *string `json:"payerId"`
	Counterparty string  `json:"counterparty"`
	Amount       float64 `json:"amount"`
	ShortTitle   string  `json:"shortTitle"`
	Comment      string  `json:"comment"`
	IsClosed     bool    `json:"isClosed"`
	CreatedBy    string  `json:"createdBy"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
}

type CreatePaymentInput struct {
	Date         int64   `json:"date"`
	RemindAt     *int64  `json:"remindAt"`
	PayerID      *string `json:"payerId"`
	Counterparty string  `json:"counterparty"`
	Amount       float64 `json:"amount"`
	ShortTitle   string  `json:"shortTitle"`
	Comment      string  `json:"comment"`
}

type PatchPaymentInput struct {
	Date         int64   `json:"date"`
	RemindAt     *int64  `json:"remindAt"`
	PayerID      *string `json:"payerId"`
	Counterparty string  `json:"counterparty"`
	Amount       float64 `json:"amount"`
	ShortTitle   string  `json:"shortTitle"`
	Comment      string  `json:"comment"`
	IsClosed     *bool   `json:"isClosed"`
}
