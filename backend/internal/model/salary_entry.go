package model

type SalaryEntry struct {
	ID              string  `json:"id"`
	Date            int64   `json:"date"`
	DealID          string  `json:"dealId"`
	DealNumberLabel string  `json:"dealNumberLabel"`
	Service         string  `json:"service"`
	Salary          float64 `json:"salary"`
	Comment         string  `json:"comment"`
	CreatedBy       string  `json:"createdBy"`
	CreatedAt       int64   `json:"createdAt"`
	UpdatedAt       int64   `json:"updatedAt"`
}

type UpsertSalaryEntryInput struct {
	Date       int64   `json:"date"`
	DealID     string  `json:"dealId"`
	Service    string  `json:"service"`
	Salary     float64 `json:"salary"`
	Comment    string  `json:"comment"`
	EmployeeID string  `json:"employeeId"`
}
