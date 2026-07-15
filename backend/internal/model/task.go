package model

type Task struct {
	ID               string       `json:"id"`
	Title            string       `json:"title"`
	Text             string       `json:"text"`
	DueAt            *int64       `json:"dueAt"`
	Status           string       `json:"status"`
	LeadID           *string      `json:"leadId,omitempty"`
	DealID           *string      `json:"dealId,omitempty"`
	ClientFirstName  string       `json:"clientFirstName"`
	ClientPatronymic string       `json:"clientPatronymic"`
	ClientPhone      string       `json:"clientPhone"`
	TrafficSource    string       `json:"trafficSource"`
	CreatedBy        string       `json:"createdBy"`
	CreatedAt        int64        `json:"createdAt"`
	UpdatedAt        int64        `json:"updatedAt"`
	Attachments      []Attachment `json:"attachments"`
	Activities       []Activity   `json:"activities"`
}

type CreateTaskInput struct {
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	DueAt     *int64  `json:"dueAt"`
	Status    string  `json:"status"`
	LeadID    *string `json:"leadId"`
	DealID    *string `json:"dealId"`
	CreatedBy string  `json:"createdBy"`
}

type UpdateTaskInput struct {
	Title    *string `json:"title"`
	Text     *string `json:"text"`
	DueAt    *int64  `json:"dueAt"`
	HasDueAt bool
}
