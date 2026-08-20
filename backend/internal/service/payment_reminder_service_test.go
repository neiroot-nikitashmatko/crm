package service

import (
	"strings"
	"testing"
	"time"

	"proclients/backend/internal/model"
)

func TestFormatAmountRub(t *testing.T) {
	cases := map[float64]string{
		0:       "0 ₽",
		50000:   "50 000 ₽",
		50000.5: "50 000,50 ₽",
		1234.56: "1 234,56 ₽",
		-10:     "-10 ₽",
	}
	for amount, want := range cases {
		if got := formatAmountRub(amount); got != want {
			t.Errorf("formatAmountRub(%v) = %q, want %q", amount, got, want)
		}
	}
}

func TestFormatPaymentReminderOmitsEmptyFields(t *testing.T) {
	payer := "ip-shmatko-nikita"
	text := formatPaymentReminder(model.Payment{
		ShortTitle:   "Налоги УСН",
		Date:         time.Date(2026, 8, 25, 0, 0, 0, 0, moscowLocation).UnixMilli(),
		Amount:       15000,
		PayerID:      &payer,
		Counterparty: "",
		Comment:      "",
	})
	if !strings.Contains(text, "Налоги УСН") {
		t.Fatalf("missing title: %s", text)
	}
	if !strings.Contains(text, "Сумма: 15 000 ₽") {
		t.Fatalf("missing amount: %s", text)
	}
	if !strings.Contains(text, "Дата: 25.08.2026") {
		t.Fatalf("missing date: %s", text)
	}
	if strings.Contains(text, "сегодня") || strings.Contains(text, "срок прошёл") {
		t.Fatalf("date should be plain: %s", text)
	}
	if !strings.Contains(text, "Плательщик: ИП Шматко Никита Сергеевич") {
		t.Fatalf("missing payer: %s", text)
	}
	if strings.Contains(text, "Контрагент:") {
		t.Fatalf("empty counterparty should be omitted: %s", text)
	}
	if strings.Contains(text, "<i>") {
		t.Fatalf("empty comment should be omitted: %s", text)
	}

	withComment := formatPaymentReminder(model.Payment{
		ShortTitle: "Оплатить Билайн",
		Date:       time.Date(2026, 8, 20, 0, 0, 0, 0, moscowLocation).UnixMilli(),
		Amount:     5456,
		Comment:    "Тест",
	})
	if !strings.Contains(withComment, "<i>Тест</i>") {
		t.Fatalf("comment should be italic: %s", withComment)
	}
}

func TestNextMoscowReminderTime(t *testing.T) {
	before := time.Date(2026, 8, 20, 9, 59, 0, 0, moscowLocation)
	got := nextMoscowReminderTime(before)
	want := time.Date(2026, 8, 20, 10, 0, 0, 0, moscowLocation)
	if !got.Equal(want) {
		t.Fatalf("before 10:00: got %v want %v", got, want)
	}

	after := time.Date(2026, 8, 20, 10, 0, 0, 0, moscowLocation)
	got = nextMoscowReminderTime(after)
	want = time.Date(2026, 8, 21, 10, 0, 0, 0, moscowLocation)
	if !got.Equal(want) {
		t.Fatalf("at 10:00: got %v want %v", got, want)
	}
}
