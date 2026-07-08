package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"proclients/backend/internal/constants"
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
	OK            bool   `json:"ok"`
	Action        string `json:"action,omitempty"`
	LeadID        string `json:"leadId,omitempty"`
	NormalizedTo  string `json:"normalizedPhone,omitempty"`
	TrafficSource string `json:"trafficSource,omitempty"`
}

func (s *BeelineIntegrationService) VerifySecret(provided string) bool {
	provided = strings.TrimSpace(provided)
	if s.webhookSecret == "" || provided == "" {
		return false
	}
	return provided == s.webhookSecret
}

func (s *BeelineIntegrationService) HandleXSIEvent(
	ctx context.Context,
	rawBody []byte,
	contentType string,
	trafficSourceHint string,
) (BeelineWebhookResult, error) {
	if strings.TrimSpace(s.createdByUserID) == "" {
		return BeelineWebhookResult{}, errors.New("BEELINE_CREATED_BY_USER_ID is required")
	}

	phone := extractCallerPhoneFromEvent(rawBody, contentType)
	if phone == "" {
		return BeelineWebhookResult{OK: true, Action: "ignored"}, nil
	}

	normalized := normalizeRUPhone(phone)
	if normalized == "" {
		return BeelineWebhookResult{OK: true, Action: "ignored"}, nil
	}

	trafficSource := resolveBeelineTrafficSource(trafficSourceHint, rawBody, contentType)

	existingID, err := s.leadRepo.FindActiveLeadIDByPhone(ctx, normalized)
	if err != nil {
		return BeelineWebhookResult{}, err
	}
	if existingID != "" {
		return BeelineWebhookResult{
			OK:            true,
			Action:        "exists",
			LeadID:        existingID,
			NormalizedTo:  normalized,
			TrafficSource: trafficSource,
		}, nil
	}

	created, err := s.leads.Create(ctx, model.CreateLeadInput{
		FirstName:     "Входящий звонок",
		Patronymic:    "",
		Phone:         normalized,
		TrafficSource: trafficSource,
		ColumnID:      "new",
		CreatedBy:     s.createdByUserID,
	})
	if err != nil {
		return BeelineWebhookResult{}, err
	}

	if s.events != nil {
		s.events.PublishLeadCreated(LeadCreatedEvent{Lead: created})
	}

	return BeelineWebhookResult{
		OK:            true,
		Action:        "created",
		LeadID:        created.ID,
		NormalizedTo:  normalized,
		TrafficSource: trafficSource,
	}, nil
}

func resolveBeelineTrafficSource(hint string, rawBody []byte, contentType string) string {
	if source := strings.TrimSpace(hint); source != "" {
		return source
	}

	if called := extractCalledPhoneFromEvent(rawBody, contentType); called != "" {
		if source := constants.BeelineTrafficSourceForPhoneDigits(phoneDigitsKey(called)); source != "" {
			return source
		}
	}

	for digits, source := range constants.BeelineTrafficSourceByPhone {
		if strings.Contains(string(rawBody), digits) {
			return source
		}
	}

	return "Билайн"
}

func phoneDigitsKey(value string) string {
	digits := make([]rune, 0, 10)
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	raw := string(digits)
	if len(raw) >= 10 {
		return raw[len(raw)-10:]
	}
	return raw
}

var ruPhoneRegex = regexp.MustCompile(`(?m)(\+?\d[\d\-\s\(\)]{9,}\d)`)

func extractCallerPhoneFromEvent(rawBody []byte, contentType string) string {
	trimmed := strings.TrimSpace(string(rawBody))
	if trimmed == "" {
		return ""
	}

	if strings.Contains(strings.ToLower(contentType), "application/json") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var payload any
		if err := json.Unmarshal(rawBody, &payload); err == nil {
			if candidate := findPhoneInJSON(payload, callerPhoneJSONKeys); candidate != "" {
				return candidate
			}
		}
	}

	match := ruPhoneRegex.FindStringSubmatch(trimmed)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func extractCalledPhoneFromEvent(rawBody []byte, contentType string) string {
	trimmed := strings.TrimSpace(string(rawBody))
	if trimmed == "" {
		return ""
	}

	if strings.Contains(strings.ToLower(contentType), "application/json") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var payload any
		if err := json.Unmarshal(rawBody, &payload); err == nil {
			if candidate := findPhoneInJSON(payload, calledPhoneJSONKeys); candidate != "" {
				return candidate
			}
		}
	}

	for _, key := range []string{"CalledNumber", "calledNumber", "DNIS", "destination"} {
		open := strings.Index(strings.ToLower(trimmed), strings.ToLower(key))
		if open == -1 {
			continue
		}
		fragment := trimmed[open:]
		if match := ruPhoneRegex.FindStringSubmatch(fragment); len(match) > 1 {
			return match[1]
		}
	}

	return ""
}

var callerPhoneJSONKeys = []string{
	"phone", "caller", "callingnumber", "from", "ani", "callingparty", "remote",
}

var calledPhoneJSONKeys = []string{
	"callednumber", "called", "to", "dnis", "destination", "terminating", "dialed",
}

func findPhoneInJSON(value any, keys []string) string {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			lower := strings.ToLower(key)
			for _, candidateKey := range keys {
				if lower == candidateKey {
					if s, ok := nested.(string); ok && strings.TrimSpace(s) != "" {
						return s
					}
				}
			}
			if candidate := findPhoneInJSON(nested, keys); candidate != "" {
				return candidate
			}
		}
	case []any:
		for _, nested := range typed {
			if candidate := findPhoneInJSON(nested, keys); candidate != "" {
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

