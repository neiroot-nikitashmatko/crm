package model

import "time"

type QuickReplySection struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	SortOrder int          `json:"sortOrder"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Replies   []QuickReply `json:"replies,omitempty"`
}

type QuickReply struct {
	ID        string    `json:"id"`
	SectionID string    `json:"sectionId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateQuickReplySectionInput struct {
	Title string `json:"title"`
}

type UpdateQuickReplySectionInput struct {
	Title string `json:"title"`
}

type CreateQuickReplyInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type UpdateQuickReplyInput struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
