package service

import (
	"sync"

	"proclients/backend/internal/model"
)

type LeadCreatedEvent struct {
	Lead model.Lead `json:"lead"`
}

type AvitoMessageEvent struct {
	LeadID      string             `json:"leadId"`
	Message     model.AvitoMessage `json:"message"`
	CreatedLead bool               `json:"createdLead,omitempty"`
}

type AvitoChatReadEvent struct {
	LeadID string `json:"leadId"`
}

type EventBus struct {
	mu               sync.Mutex
	leadCreatedSubs  map[chan LeadCreatedEvent]struct{}
	avitoMessageSubs map[chan AvitoMessageEvent]struct{}
	avitoChatReadSubs map[chan AvitoChatReadEvent]struct{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		leadCreatedSubs:   make(map[chan LeadCreatedEvent]struct{}),
		avitoMessageSubs:  make(map[chan AvitoMessageEvent]struct{}),
		avitoChatReadSubs: make(map[chan AvitoChatReadEvent]struct{}),
	}
}

func (b *EventBus) SubscribeLeadCreated() (ch chan LeadCreatedEvent, unsubscribe func()) {
	ch = make(chan LeadCreatedEvent, 8)
	b.mu.Lock()
	b.leadCreatedSubs[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.leadCreatedSubs[ch]; ok {
			delete(b.leadCreatedSubs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *EventBus) PublishLeadCreated(event LeadCreatedEvent) {
	b.mu.Lock()
	subs := make([]chan LeadCreatedEvent, 0, len(b.leadCreatedSubs))
	for ch := range b.leadCreatedSubs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// Drop if subscriber is slow.
		}
	}
}

func (b *EventBus) SubscribeAvitoMessage() (ch chan AvitoMessageEvent, unsubscribe func()) {
	ch = make(chan AvitoMessageEvent, 16)
	b.mu.Lock()
	b.avitoMessageSubs[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.avitoMessageSubs[ch]; ok {
			delete(b.avitoMessageSubs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *EventBus) PublishAvitoMessage(event AvitoMessageEvent) {
	b.mu.Lock()
	subs := make([]chan AvitoMessageEvent, 0, len(b.avitoMessageSubs))
	for ch := range b.avitoMessageSubs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (b *EventBus) SubscribeAvitoChatRead() (ch chan AvitoChatReadEvent, unsubscribe func()) {
	ch = make(chan AvitoChatReadEvent, 16)
	b.mu.Lock()
	b.avitoChatReadSubs[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if _, ok := b.avitoChatReadSubs[ch]; ok {
			delete(b.avitoChatReadSubs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *EventBus) PublishAvitoChatRead(event AvitoChatReadEvent) {
	b.mu.Lock()
	subs := make([]chan AvitoChatReadEvent, 0, len(b.avitoChatReadSubs))
	for ch := range b.avitoChatReadSubs {
		subs = append(subs, ch)
	}
	b.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}
