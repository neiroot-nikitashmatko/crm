package service

import (
	"sync"

	"proclients/backend/internal/model"
)

type LeadCreatedEvent struct {
	Lead model.Lead `json:"lead"`
}

type EventBus struct {
	mu             sync.Mutex
	leadCreatedSubs map[chan LeadCreatedEvent]struct{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		leadCreatedSubs: make(map[chan LeadCreatedEvent]struct{}),
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

