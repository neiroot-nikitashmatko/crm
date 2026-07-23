package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"proclients/backend/internal/avito"
	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type AvitoIntegrationService struct {
	client          *avito.Client
	avitoRepo       *repository.AvitoChatRepository
	leads           *LeadService
	leadRepo        *repository.LeadRepository
	events          *EventBus
	webhookSecret   string
	createdByUserID string
	debugLogging    bool
}

type AvitoWebhookResult struct {
	OK        bool   `json:"ok"`
	Action    string `json:"action"`
	LeadID    string `json:"leadId,omitempty"`
	ChatID    string `json:"chatId,omitempty"`
	MessageID string `json:"messageId,omitempty"`
	Direction string `json:"direction,omitempty"`
	Inserted  bool   `json:"inserted,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

func NewAvitoIntegrationService(
	client *avito.Client,
	avitoRepo *repository.AvitoChatRepository,
	leads *LeadService,
	leadRepo *repository.LeadRepository,
	events *EventBus,
	webhookSecret string,
	createdByUserID string,
	debugLogging bool,
) *AvitoIntegrationService {
	return &AvitoIntegrationService{
		client:          client,
		avitoRepo:       avitoRepo,
		leads:           leads,
		leadRepo:        leadRepo,
		events:          events,
		webhookSecret:   strings.TrimSpace(webhookSecret),
		createdByUserID: strings.TrimSpace(createdByUserID),
		debugLogging:    debugLogging,
	}
}

func (s *AvitoIntegrationService) Enabled() bool {
	return s != nil && s.client != nil && s.client.Enabled()
}

func (s *AvitoIntegrationService) VerifySecret(provided string) bool {
	if s.webhookSecret == "" {
		return false
	}
	return strings.TrimSpace(provided) == s.webhookSecret
}

func (s *AvitoIntegrationService) SubscribeWebhook(ctx context.Context, publicURL string) error {
	if !s.Enabled() {
		return fmt.Errorf("avito integration is not configured")
	}
	return s.client.SubscribeWebhook(ctx, publicURL)
}

type AvitoLeadChatBundle struct {
	Linked   bool                 `json:"linked"`
	Chat     *model.AvitoChat     `json:"chat,omitempty"`
	Messages []model.AvitoMessage `json:"messages"`
}

func (s *AvitoIntegrationService) GetChatForLead(ctx context.Context, leadID string) (model.AvitoChat, error) {
	chat, err := s.avitoRepo.GetByLeadID(ctx, leadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.AvitoChat{}, fmt.Errorf("avito chat not linked to lead")
		}
		return model.AvitoChat{}, err
	}
	return chat, nil
}

func (s *AvitoIntegrationService) ListChats(ctx context.Context, userID string) ([]model.AvitoChat, error) {
	items, err := s.avitoRepo.ListChats(ctx, userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []model.AvitoChat{}, nil
	}
	return items, nil
}

func (s *AvitoIntegrationService) GetLeadChatBundle(ctx context.Context, leadID string) (AvitoLeadChatBundle, error) {
	leadID = strings.TrimSpace(leadID)
	if leadID == "" {
		return AvitoLeadChatBundle{}, fmt.Errorf("leadId is required")
	}

	chat, err := s.avitoRepo.GetByLeadID(ctx, leadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AvitoLeadChatBundle{Linked: false, Messages: []model.AvitoMessage{}}, nil
		}
		return AvitoLeadChatBundle{}, err
	}

	messages, listErr := s.avitoRepo.ListMessagesByChatID(ctx, chat.ChatID)
	if listErr != nil {
		return AvitoLeadChatBundle{}, listErr
	}
	if messages == nil {
		messages = []model.AvitoMessage{}
	}

	// Don't block chat open on Avito API. Webhooks keep history up to date.
	// Pull from Avito only when local history is empty (first open / missed sync).
	if s.Enabled() && len(messages) == 0 {
		if _, syncErr := s.SyncChatMessages(ctx, chat.ChatID); syncErr != nil {
			s.logDebug("sync messages for lead %s failed: %v", leadID, syncErr)
		} else {
			messages, listErr = s.avitoRepo.ListMessagesByChatID(ctx, chat.ChatID)
			if listErr != nil {
				return AvitoLeadChatBundle{}, listErr
			}
			if messages == nil {
				messages = []model.AvitoMessage{}
			}
		}
	}

	return AvitoLeadChatBundle{
		Linked:   true,
		Chat:     &chat,
		Messages: messages,
	}, nil
}

func (s *AvitoIntegrationService) ListMessagesForLead(ctx context.Context, leadID string) ([]model.AvitoMessage, error) {
	messages, err := s.avitoRepo.ListMessagesByLeadID(ctx, leadID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []model.AvitoMessage{}, nil
		}
		return nil, err
	}
	return messages, nil
}

func (s *AvitoIntegrationService) SendMessageToLead(ctx context.Context, leadID, text string, files []avito.UploadFile) ([]model.AvitoMessage, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("avito integration is not configured")
	}
	text = strings.TrimSpace(text)
	if text == "" && len(files) == 0 {
		return nil, fmt.Errorf("text or image is required")
	}

	chat, err := s.avitoRepo.GetByLeadID(ctx, leadID)
	if err != nil {
		return nil, fmt.Errorf("avito chat not linked to lead")
	}

	sentTexts := make([]string, 0, 1+len(files))
	if text != "" {
		if sendErr := s.client.SendTextMessage(ctx, chat.ChatID, text); sendErr != nil {
			return nil, sendErr
		}
		sentTexts = append(sentTexts, text)
	}

	for _, file := range files {
		if !isAllowedAvitoImage(file) {
			return nil, fmt.Errorf("Авито принимает только изображения (JPEG, PNG, GIF, WEBP, BMP, HEIC)")
		}
		imageID, sizes, uploadErr := s.client.UploadImage(ctx, file)
		if uploadErr != nil {
			return nil, uploadErr
		}
		if sendErr := s.client.SendImageMessage(ctx, chat.ChatID, imageID); sendErr != nil {
			return nil, sendErr
		}
		preview := avito.PreferredImageURL(sizes)
		if preview == "" {
			preview = "[Изображение]"
		}
		sentTexts = append(sentTexts, preview)
	}

	if _, syncErr := s.SyncChatMessages(ctx, chat.ChatID); syncErr != nil {
		s.logDebug("sync after send for %s failed: %v", chat.ChatID, syncErr)
	}

	messages, listErr := s.avitoRepo.ListMessagesByChatID(ctx, chat.ChatID)
	if listErr != nil {
		return nil, listErr
	}

	result := make([]model.AvitoMessage, 0, len(sentTexts))
	for i := len(messages) - 1; i >= 0 && len(result) < len(sentTexts); i-- {
		item := messages[i]
		if item.Direction != "outgoing" {
			continue
		}
		result = append([]model.AvitoMessage{item}, result...)
	}
	if len(result) > 0 {
		for _, item := range result {
			s.publishAvitoMessage(leadID, item, false)
		}
		return result, nil
	}

	// Fallback if Avito API has not returned the messages yet.
	now := time.Now().UTC()
	authorID := s.client.UserID()
	for idx, value := range sentTexts {
		messageType := "text"
		if strings.HasPrefix(value, "http") || value == "[Изображение]" {
			messageType = "image"
		}
		inserted, _, insertErr := s.avitoRepo.InsertMessage(ctx, model.InsertAvitoMessageInput{
			ChatID:      chat.ChatID,
			MessageID:   fmt.Sprintf("out-%d-%d", now.UnixNano(), idx),
			Direction:   "outgoing",
			MessageType: messageType,
			Text:        value,
			AuthorID:    &authorID,
			SentAt:      now,
		})
		if insertErr != nil {
			return result, insertErr
		}
		result = append(result, inserted)
		s.publishAvitoMessage(leadID, inserted, false)
	}
	return result, nil
}

func isAllowedAvitoImage(file avito.UploadFile) bool {
	contentType := strings.ToLower(strings.TrimSpace(file.ContentType))
	switch contentType {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/bmp", "image/heic", "image/heif":
		return true
	}
	name := strings.ToLower(file.Filename)
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".heic", ".heif"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func (s *AvitoIntegrationService) publishAvitoMessage(leadID string, message model.AvitoMessage, createdLead bool) {
	if s.events == nil || strings.TrimSpace(leadID) == "" || strings.TrimSpace(message.ID) == "" {
		return
	}
	s.events.PublishAvitoMessage(AvitoMessageEvent{
		LeadID:      leadID,
		Message:     message,
		CreatedLead: createdLead,
	})
}

func (s *AvitoIntegrationService) SyncChatMessages(ctx context.Context, chatID string) (int, error) {
	if !s.Enabled() {
		return 0, fmt.Errorf("avito integration is not configured")
	}

	response, err := s.client.GetMessages(ctx, chatID, 50, 0)
	if err != nil {
		return 0, err
	}

	insertedCount := 0
	for _, item := range response.Messages {
		direction := "incoming"
		if item.Direction == "out" || item.AuthorID == s.client.UserID() {
			direction = "outgoing"
		}
		authorID := item.AuthorID
		sentAt := time.Unix(item.Created, 0).UTC()
		_, inserted, insertErr := s.avitoRepo.InsertMessage(ctx, model.InsertAvitoMessageInput{
			ChatID:      chatID,
			MessageID:   item.ID,
			Direction:   direction,
			MessageType: fallback(item.Type, "text"),
			Text:        avito.MessageText(item),
			AuthorID:    &authorID,
			SentAt:      sentAt,
		})
		if insertErr != nil {
			return insertedCount, insertErr
		}
		if inserted {
			insertedCount++
		}
	}
	return insertedCount, nil
}

type avitoWebhookEnvelope struct {
	ID        string          `json:"id"`
	Version   string          `json:"version"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type avitoWebhookPayload struct {
	Type  string              `json:"type"`
	Value avitoWebhookMessage `json:"value"`
}

type avitoWebhookMessage struct {
	ID       string `json:"id"`
	ChatID   string `json:"chat_id"`
	UserID   int64  `json:"user_id"`
	AuthorID int64  `json:"author_id"`
	Created  int64  `json:"created"`
	Type     string `json:"type"`
	ItemID   *int64 `json:"item_id"`
	ChatType string `json:"chat_type"`
	Content  struct {
		Text  string `json:"text"`
		Image *struct {
			Sizes map[string]string `json:"sizes"`
		} `json:"image"`
	} `json:"content"`
}

func (s *AvitoIntegrationService) HandleWebhook(ctx context.Context, rawBody []byte) (AvitoWebhookResult, error) {
	s.logDebug("webhook body: %s", truncateAvitoForLog(string(rawBody), 2000))

	if strings.TrimSpace(s.createdByUserID) == "" {
		return AvitoWebhookResult{OK: true, Action: "ignored", Detail: "AVITO_CREATED_BY_USER_ID is not configured"}, nil
	}
	if !s.Enabled() {
		return AvitoWebhookResult{OK: true, Action: "ignored", Detail: "avito client is not configured"}, nil
	}

	var envelope avitoWebhookEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return AvitoWebhookResult{}, fmt.Errorf("invalid webhook json: %w", err)
	}

	var payload avitoWebhookPayload
	if len(envelope.Payload) > 0 {
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return AvitoWebhookResult{}, fmt.Errorf("invalid webhook payload: %w", err)
		}
	} else {
		// fallback: body itself is payload
		if err := json.Unmarshal(rawBody, &payload); err != nil {
			return AvitoWebhookResult{}, fmt.Errorf("invalid webhook payload: %w", err)
		}
	}

	if strings.TrimSpace(payload.Type) != "" && payload.Type != "message" {
		return AvitoWebhookResult{OK: true, Action: "ignored", Detail: "unsupported payload type: " + payload.Type}, nil
	}

	msg := payload.Value
	if strings.TrimSpace(msg.ChatID) == "" || strings.TrimSpace(msg.ID) == "" {
		return AvitoWebhookResult{OK: true, Action: "ignored", Detail: "missing chat_id or message id"}, nil
	}

	direction := "incoming"
	if msg.AuthorID == s.client.UserID() {
		direction = "outgoing"
	}

	leadID, createdLead, err := s.ensureLeadForChat(ctx, msg.ChatID, msg.ItemID, direction)
	if err != nil {
		return AvitoWebhookResult{}, err
	}

	authorID := msg.AuthorID
	sentAt := time.Now().UTC()
	if msg.Created > 0 {
		sentAt = time.Unix(msg.Created, 0).UTC()
	}

	text := strings.TrimSpace(msg.Content.Text)
	if text == "" && msg.Content.Image != nil {
		text = avito.PreferredImageURL(msg.Content.Image.Sizes)
	}
	if text == "" && msg.Type != "" && msg.Type != "text" {
		text = "[" + msg.Type + "]"
	}

	stored, inserted, insertErr := s.avitoRepo.InsertMessage(ctx, model.InsertAvitoMessageInput{
		ChatID:      msg.ChatID,
		MessageID:   msg.ID,
		Direction:   direction,
		MessageType: fallback(msg.Type, "text"),
		Text:        text,
		AuthorID:    &authorID,
		SentAt:      sentAt,
	})
	if insertErr != nil {
		return AvitoWebhookResult{}, insertErr
	}

	action := "message"
	if createdLead {
		action = "lead_created"
	} else if !inserted {
		action = "duplicate"
	}

	if inserted {
		s.publishAvitoMessage(leadID, stored, createdLead)
	}

	return AvitoWebhookResult{
		OK:        true,
		Action:    action,
		LeadID:    leadID,
		ChatID:    msg.ChatID,
		MessageID: msg.ID,
		Direction: direction,
		Inserted:  inserted,
	}, nil
}

func (s *AvitoIntegrationService) ensureLeadForChat(
	ctx context.Context,
	chatID string,
	itemID *int64,
	direction string,
) (leadID string, created bool, err error) {
	existing, getErr := s.avitoRepo.GetByChatID(ctx, chatID)
	if getErr != nil && getErr != pgx.ErrNoRows {
		return "", false, getErr
	}

	if getErr == nil {
		lead, leadErr := s.leadRepo.GetByID(ctx, existing.LeadID)
		if leadErr == nil && lead.ColumnID != "failed" {
			return existing.LeadID, false, nil
		}
		if leadErr == nil && lead.ColumnID == "failed" {
			// New CRM lead only when the client writes again — not on our outgoing.
			if direction != "incoming" {
				return existing.LeadID, false, nil
			}
			s.logDebug("chat %s linked to failed lead %s, recreating", chatID, existing.LeadID)
		} else {
			// Linked lead was soft-deleted or missing — create a new one and re-bind.
			s.logDebug("chat %s linked to inactive lead %s, recreating", chatID, existing.LeadID)
		}
	}

	nickname := "Пользователь Авито"
	var peerUserID *int64
	avatarURL := ""
	itemTitle := ""
	var resolvedItemID *int64 = itemID

	if getErr == nil {
		if strings.TrimSpace(existing.PeerNickname) != "" {
			nickname = existing.PeerNickname
		}
		peerUserID = existing.PeerUserID
		avatarURL = existing.PeerAvatarURL
		if existing.ItemID != nil {
			resolvedItemID = existing.ItemID
		}
		itemTitle = existing.ItemTitle
	}

	if chat, chatErr := s.client.GetChat(ctx, chatID); chatErr == nil {
		peerID, peerName, peerAvatar := s.client.PeerFromChat(chat)
		if peerName != "" {
			nickname = peerName
		}
		if peerID > 0 {
			peerUserID = &peerID
		}
		avatarURL = peerAvatar
		if chat.Context.Value.ID > 0 {
			id := chat.Context.Value.ID
			resolvedItemID = &id
			itemTitle = strings.TrimSpace(chat.Context.Value.Title)
		}
	} else {
		s.logDebug("get chat %s failed: %v", chatID, chatErr)
	}

	lead, createErr := s.leads.Create(ctx, model.CreateLeadInput{
		FirstName:     nickname,
		Patronymic:    "",
		Phone:         "",
		TrafficSource: "Авито Чат",
		ColumnID:      "new",
		CreatedBy:     s.createdByUserID,
	})
	if createErr != nil {
		return "", false, createErr
	}

	if _, upsertErr := s.avitoRepo.UpsertChat(ctx, model.UpsertAvitoChatInput{
		ChatID:        chatID,
		LeadID:        lead.ID,
		PeerUserID:    peerUserID,
		PeerNickname:  nickname,
		PeerAvatarURL: avatarURL,
		ItemID:        resolvedItemID,
		ItemTitle:     itemTitle,
	}); upsertErr != nil {
		return "", false, upsertErr
	}

	if s.events != nil {
		s.events.PublishLeadCreated(LeadCreatedEvent{Lead: lead})
	}

	// Best-effort history sync for new chats.
	if _, syncErr := s.SyncChatMessages(ctx, chatID); syncErr != nil {
		s.logDebug("sync messages for %s failed: %v", chatID, syncErr)
	}

	return lead.ID, true, nil
}

func (s *AvitoIntegrationService) logDebug(format string, args ...any) {
	if !s.debugLogging {
		return
	}
	log.Printf("[avito] "+format, args...)
}

func fallback(value, def string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return def
	}
	return value
}

func truncateAvitoForLog(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "…"
}
