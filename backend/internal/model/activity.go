package model

const (
	ActivityEntityDeal = "deal"
	ActivityEntityTask = "task"

	ActivityTypeSystem  = "system"
	ActivityTypeComment = "comment"
)

type Activity struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"createdAt"`
}

type CreateActivityInput struct {
	Text string `json:"text"`
	Type string `json:"type"`
}
