package model

import "time"

type AvitoChat struct {
	ID             string    `json:"id"`
	ChatID         string    `json:"chatId"`
	LeadID         string    `json:"leadId"`
	PeerUserID     *int64    `json:"peerUserId,omitempty"`
	PeerNickname   string    `json:"peerNickname"`
	PeerAvatarURL  string    `json:"peerAvatarUrl"`
	ItemID         *int64    `json:"itemId,omitempty"`
	ItemTitle      string    `json:"itemTitle"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	UnreadCount    int       `json:"unreadCount"`
}

type AvitoMessage struct {
	ID          string    `json:"id"`
	ChatID      string    `json:"chatId"`
	MessageID   string    `json:"messageId"`
	Direction   string    `json:"direction"`
	MessageType string    `json:"messageType"`
	Text        string    `json:"text"`
	AuthorID    *int64    `json:"authorId,omitempty"`
	SentAt      time.Time `json:"sentAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UpsertAvitoChatInput struct {
	ChatID        string
	LeadID        string
	PeerUserID    *int64
	PeerNickname  string
	PeerAvatarURL string
	ItemID        *int64
	ItemTitle     string
}

type InsertAvitoMessageInput struct {
	ChatID      string
	MessageID   string
	Direction   string
	MessageType string
	Text        string
	AuthorID    *int64
	SentAt      time.Time
}
