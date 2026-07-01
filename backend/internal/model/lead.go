package model

type Lead struct {
	ID             string         `json:"id"`
	LeadNumber     int64          `json:"leadNumber"`
	FirstName      string         `json:"firstName"`
	Patronymic     string         `json:"patronymic"`
	Phone          string         `json:"phone"`
	TrafficSource  string         `json:"trafficSource"`
	ColumnID       string         `json:"columnId"`
	LeadComments   string         `json:"leadComments"`
	FailureReason  string         `json:"failureReason"`
	CreatedBy      string         `json:"createdBy"`
	CreatedAt      int64          `json:"createdAt"`
	UpdatedAt      int64          `json:"updatedAt"`
	PickupDelivery PickupDelivery `json:"pickupDelivery"`
	Products       []DealProduct  `json:"products"`
	Production     DealProduction `json:"production"`
}

type CreateLeadInput struct {
	FirstName     string `json:"firstName"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	TrafficSource string `json:"trafficSource"`
	ColumnID      string `json:"columnId"`
	CreatedBy     string `json:"createdBy"`
}
