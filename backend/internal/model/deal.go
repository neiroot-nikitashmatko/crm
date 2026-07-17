package model

type DealProduct struct {
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type DealProduction struct {
	Nomenclature string `json:"nomenclature"`
	DueAt        *int64 `json:"dueAt"`
	Employee     string `json:"employee"`
}

type PickupDelivery struct {
	PickupAddress   string `json:"pickupAddress"`
	PickupDate      *int64 `json:"pickupDate"`
	DeliveryAddress string `json:"deliveryAddress"`
	DeliveryDate    *int64 `json:"deliveryDate"`
	Courier         string `json:"courier"`
}

type Deal struct {
	ID              string         `json:"id"`
	LeadID          *string        `json:"leadId,omitempty"`
	DealNumber      int64          `json:"dealNumber"`
	FirstName       string         `json:"firstName"`
	Patronymic      string         `json:"patronymic"`
	Phone           string         `json:"phone"`
	TrafficSource   string         `json:"trafficSource"`
	Status          string         `json:"status"`
	TotalAmount     float64        `json:"totalAmount"`
	DealComments    string         `json:"dealComments"`
	FailureReason   string         `json:"failureReason"`
	CreatedBy       string         `json:"createdBy"`
	CreatedByName   string         `json:"createdByName"`
	CreatedAt       int64          `json:"createdAt"`
	UpdatedAt       int64          `json:"updatedAt"`
	ProductionDueAt *int64         `json:"productionDueAt"`
	Production      DealProduction `json:"production"`
	PickupDelivery  PickupDelivery `json:"pickupDelivery"`
	Products        []DealProduct  `json:"products"`
	Attachments     []Attachment   `json:"attachments"`
	Activities      []Activity     `json:"activities"`
}

type CreateDealFromLeadInput struct {
	LeadID         string          `json:"leadId"`
	CreatedBy      string          `json:"createdBy"`
	Products       []DealProduct   `json:"products"`
	Production     DealProduction  `json:"production"`
	PickupDelivery *PickupDelivery `json:"pickupDelivery,omitempty"`
}
