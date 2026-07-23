package service

import (
	"context"
	"errors"
	"strings"

	"proclients/backend/internal/repository"
)

type NotificationSummary struct {
	NewLeadsCount    int `json:"newLeadsCount"`
	UnreadChatsCount int `json:"unreadChatsCount"`
}

type NotificationService struct {
	leads *repository.LeadRepository
	avito *repository.AvitoChatRepository
}

func NewNotificationService(
	leads *repository.LeadRepository,
	avito *repository.AvitoChatRepository,
) *NotificationService {
	return &NotificationService{leads: leads, avito: avito}
}

func (s *NotificationService) Summary(ctx context.Context, userID string) (NotificationSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return NotificationSummary{}, errors.New("userId is required")
	}

	newLeads, err := s.leads.CountByColumn(ctx, "new")
	if err != nil {
		return NotificationSummary{}, err
	}
	unreadChats, err := s.avito.CountUnreadChatsForUser(ctx, userID)
	if err != nil {
		return NotificationSummary{}, err
	}

	return NotificationSummary{
		NewLeadsCount:    newLeads,
		UnreadChatsCount: unreadChats,
	}, nil
}

func (s *NotificationService) MarkAvitoChatRead(ctx context.Context, userID string, leadID string) error {
	userID = strings.TrimSpace(userID)
	leadID = strings.TrimSpace(leadID)
	if userID == "" {
		return errors.New("userId is required")
	}
	if leadID == "" {
		return errors.New("leadId is required")
	}
	return s.avito.MarkChatReadByLeadID(ctx, userID, leadID)
}
