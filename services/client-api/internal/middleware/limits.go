package middleware

import (
	"sync"
	"time"
)

type bucket struct {
	tokens  float64
	updated time.Time
}
type Limits struct {
	mu          sync.Mutex
	clients     map[string]bucket
	rate, burst int
	slots       chan struct{}
}

func New(rate, burst, concurrent int) *Limits {
	return &Limits{clients: map[string]bucket{}, rate: rate, burst: burst, slots: make(chan struct{}, concurrent)}
}
func (l *Limits) Enter(id string) (func(), bool) {
	l.mu.Lock()
	now := time.Now()
	b, ok := l.clients[id]
	if !ok {
		if len(l.clients) >= 1000 {
			for k, v := range l.clients {
				if now.Sub(v.updated) > time.Minute {
					delete(l.clients, k)
				}
			}
			if len(l.clients) >= 1000 {
				l.mu.Unlock()
				return nil, false
			}
		}
		b = bucket{tokens: float64(l.burst), updated: now}
	}
	b.tokens = min(float64(l.burst), b.tokens+now.Sub(b.updated).Seconds()*float64(l.rate))
	b.updated = now
	allowed := b.tokens >= 1
	if allowed {
		b.tokens--
	}
	l.clients[id] = b
	l.mu.Unlock()
	if !allowed {
		return nil, false
	}
	select {
	case l.slots <- struct{}{}:
		return func() { <-l.slots }, true
	default:
		return nil, false
	}
}
