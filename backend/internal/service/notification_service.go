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
	leads  *repository.LeadRepository
	avito  *repository.AvitoChatRepository
	events *EventBus
}

func NewNotificationService(
	leads *repository.LeadRepository,
	avito *repository.AvitoChatRepository,
	events *EventBus,
) *NotificationService {
	return &NotificationService{leads: leads, avito: avito, events: events}
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
	unreadChats, err := s.avito.CountUnreadChats(ctx)
	if err != nil {
		return NotificationSummary{}, err
	}

	return NotificationSummary{
		NewLeadsCount:    newLeads,
		UnreadChatsCount: unreadChats,
	}, nil
}

func (s *NotificationService) MarkAvitoChatRead(ctx context.Context, leadID string) error {
	leadID = strings.TrimSpace(leadID)
	if leadID == "" {
		return errors.New("leadId is required")
	}
	if err := s.avito.MarkChatReadByLeadID(ctx, leadID); err != nil {
		return err
	}
	if s.events != nil {
		s.events.PublishAvitoChatRead(AvitoChatReadEvent{LeadID: leadID})
	}
	return nil
}
