package service

import (
	"context"
	"fmt"
	"html"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"proclients/backend/internal/constants"
	"proclients/backend/internal/max"
	"proclients/backend/internal/model"
	"proclients/backend/internal/repository"
)

const (
	paymentReminderHourMSK = 10
	maxMessagePause        = 600 * time.Millisecond
	maxMessageTextLimit    = 4000
)

type PaymentReminderService struct {
	repo   *repository.PaymentRepository
	client *max.Client
	mu     sync.Mutex
}

func NewPaymentReminderService(repo *repository.PaymentRepository, client *max.Client) *PaymentReminderService {
	return &PaymentReminderService{repo: repo, client: client}
}

func (s *PaymentReminderService) Enabled() bool {
	return s != nil && s.client != nil && s.client.Enabled()
}

func (s *PaymentReminderService) Start(ctx context.Context) {
	if !s.Enabled() {
		log.Printf("max payment reminders disabled (missing MAX_BOT_TOKEN or MAX_CHAT_ID)")
		return
	}

	log.Printf("max payment reminders enabled (daily %02d:00 Europe/Moscow)", paymentReminderHourMSK)
	go s.loop(ctx)
}

func (s *PaymentReminderService) loop(ctx context.Context) {
	s.catchUpIfNeeded(ctx)

	for {
		wait := time.Until(nextMoscowReminderTime(time.Now()))
		if wait < time.Second {
			wait = time.Second
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := s.SendDueReminders(ctx); err != nil {
				log.Printf("max payment reminders: %v", err)
			}
		}
	}
}

func (s *PaymentReminderService) catchUpIfNeeded(ctx context.Context) {
	now := time.Now().In(moscowLocation)
	todayTen := time.Date(now.Year(), now.Month(), now.Day(), paymentReminderHourMSK, 0, 0, 0, moscowLocation)
	if now.Before(todayTen) {
		return
	}
	if err := s.SendDueReminders(ctx); err != nil {
		log.Printf("max payment reminders catch-up: %v", err)
	}
}

func (s *PaymentReminderService) SendDueReminders(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().In(moscowLocation)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, moscowLocation)

	items, err := s.repo.ListDueReminders(ctx, today)
	if err != nil {
		return fmt.Errorf("list due reminders: %w", err)
	}
	if len(items) == 0 {
		log.Printf("max payment reminders: nothing due")
		return nil
	}

	sent := 0
	for i, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(maxMessagePause):
			}
		}

		text := formatPaymentReminder(item)
		if err := s.client.SendFormattedText(ctx, text, "html"); err != nil {
			log.Printf("max payment reminders: send %s failed: %v", item.ID, err)
			continue
		}
		if err := s.repo.MarkReminderSent(ctx, item.ID); err != nil {
			log.Printf("max payment reminders: mark sent %s failed: %v", item.ID, err)
			continue
		}
		sent++
	}

	log.Printf("max payment reminders: sent %d of %d", sent, len(items))
	return nil
}

func nextMoscowReminderTime(now time.Time) time.Time {
	local := now.In(moscowLocation)
	next := time.Date(local.Year(), local.Month(), local.Day(), paymentReminderHourMSK, 0, 0, 0, moscowLocation)
	if !local.Before(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func formatPaymentReminder(item model.Payment) string {
	title := strings.TrimSpace(item.ShortTitle)
	if title == "" {
		title = "Оплата"
	}

	lines := []string{
		html.EscapeString(title),
		"",
		"Дата: " + formatMoscowDate(item.Date),
		"Сумма: " + html.EscapeString(formatAmountRub(item.Amount)),
	}

	if item.PayerID != nil {
		if label := strings.TrimSpace(constants.PaymentPayerLabel(*item.PayerID)); label != "" {
			lines = append(lines, "Плательщик: "+html.EscapeString(label))
		}
	}
	if counterparty := strings.TrimSpace(item.Counterparty); counterparty != "" {
		lines = append(lines, "Контрагент: "+html.EscapeString(counterparty))
	}
	if comment := strings.TrimSpace(item.Comment); comment != "" {
		lines = append(lines, "", formatReminderComment(comment))
	}

	text := strings.Join(lines, "\n")
	if utf8.RuneCountInString(text) <= maxMessageTextLimit {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxMessageTextLimit])
}

func formatReminderComment(comment string) string {
	escaped := html.EscapeString(comment)
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")
	return "<i>" + escaped + "</i>"
}

func formatMoscowDate(ms int64) string {
	return time.UnixMilli(ms).In(moscowLocation).Format("02.01.2006")
}

func formatAmountRub(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	cents := int64(math.Round(amount * 100))
	rubles := cents / 100
	kopecks := cents % 100

	formatted := formatIntGrouped(rubles)
	if kopecks != 0 {
		formatted = fmt.Sprintf("%s,%02d", formatted, kopecks)
	}
	if negative {
		formatted = "-" + formatted
	}
	return formatted + " ₽"
}

func formatIntGrouped(value int64) string {
	raw := strconv.FormatInt(value, 10)
	n := len(raw)
	if n <= 3 {
		return raw
	}

	var b strings.Builder
	prefix := n % 3
	if prefix == 0 {
		prefix = 3
	}
	b.WriteString(raw[:prefix])
	for i := prefix; i < n; i += 3 {
		b.WriteByte(' ')
		b.WriteString(raw[i : i+3])
	}
	return b.String()
}
