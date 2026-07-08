package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

type BeelineIntegrationService struct {
	leads           *LeadService
	leadRepo        *repository.LeadRepository
	events          *EventBus
	webhookSecret   string
	createdByUserID string
}

func NewBeelineIntegrationService(
	leads *LeadService,
	leadRepo *repository.LeadRepository,
	events *EventBus,
	webhookSecret string,
	createdByUserID string,
) *BeelineIntegrationService {
	return &BeelineIntegrationService{
		leads:           leads,
		leadRepo:        leadRepo,
		events:          events,
		webhookSecret:   strings.TrimSpace(webhookSecret),
		createdByUserID: strings.TrimSpace(createdByUserID),
	}
}

type BeelineWebhookResult struct {
	OK           bool   `json:"ok"`
	Action       string `json:"action,omitempty"`
	LeadID       string `json:"leadId,omitempty"`
	NormalizedTo string `json:"normalizedPhone,omitempty"`
}

func (s *BeelineIntegrationService) VerifySecret(provided string) bool {
	provided = strings.TrimSpace(provided)
	if s.webhookSecret == "" || provided == "" {
		return false
	}
	return provided == s.webhookSecret
}

func (s *BeelineIntegrationService) HandleXSIEvent(ctx context.Context, rawBody []byte, contentType string) (BeelineWebhookResult, error) {
	if strings.TrimSpace(s.createdByUserID) == "" {
		return BeelineWebhookResult{}, errors.New("BEELINE_CREATED_BY_USER_ID is required")
	}

	phone := extractPhoneFromEvent(rawBody, contentType)
	if phone == "" {
		return BeelineWebhookResult{OK: true, Action: "ignored"}, nil
	}

	normalized := normalizeRUPhone(phone)
	if normalized == "" {
		return BeelineWebhookResult{OK: true, Action: "ignored"}, nil
	}

	existingID, err := s.leadRepo.FindActiveLeadIDByPhone(ctx, normalized)
	if err != nil {
		return BeelineWebhookResult{}, err
	}
	if existingID != "" {
		return BeelineWebhookResult{OK: true, Action: "exists", LeadID: existingID, NormalizedTo: normalized}, nil
	}

	created, err := s.leads.Create(ctx, model.CreateLeadInput{
		FirstName:     "Входящий звонок",
		Patronymic:    "",
		Phone:         normalized,
		TrafficSource: "Билайн",
		ColumnID:      "new",
		CreatedBy:     s.createdByUserID,
	})
	if err != nil {
		return BeelineWebhookResult{}, err
	}

	if s.events != nil {
		s.events.PublishLeadCreated(LeadCreatedEvent{Lead: created})
	}

	return BeelineWebhookResult{OK: true, Action: "created", LeadID: created.ID, NormalizedTo: normalized}, nil
}

var ruPhoneRegex = regexp.MustCompile(`(?m)(\+?\d[\d\-\s\(\)]{9,}\d)`)

func extractPhoneFromEvent(rawBody []byte, contentType string) string {
	trimmed := strings.TrimSpace(string(rawBody))
	if trimmed == "" {
		return ""
	}

	// Try JSON first (many integrations send JSON).
	if strings.Contains(strings.ToLower(contentType), "application/json") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var payload any
		if err := json.Unmarshal(rawBody, &payload); err == nil {
			if candidate := findPhoneInJSON(payload); candidate != "" {
				return candidate
			}
		}
	}

	// Fallback: extract first phone-like token from raw text (XML/other).
	match := ruPhoneRegex.FindStringSubmatch(trimmed)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func findPhoneInJSON(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			lower := strings.ToLower(key)
			if lower == "phone" || lower == "caller" || lower == "callingnumber" || lower == "from" || lower == "ani" {
				if s, ok := nested.(string); ok && strings.TrimSpace(s) != "" {
					return s
				}
			}
			if candidate := findPhoneInJSON(nested); candidate != "" {
				return candidate
			}
		}
	case []any:
		for _, nested := range typed {
			if candidate := findPhoneInJSON(nested); candidate != "" {
				return candidate
			}
		}
	case string:
		return typed
	default:
		return ""
	}
	return ""
}

func normalizeRUPhone(value string) string {
	digits := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) == 0 {
		return ""
	}

	raw := string(digits)
	switch {
	case len(raw) == 11 && strings.HasPrefix(raw, "8"):
		return "+7" + raw[1:]
	case len(raw) == 11 && strings.HasPrefix(raw, "7"):
		return "+7" + raw[1:]
	case len(raw) == 10:
		return "+7" + raw
	default:
		return ""
	}
}

