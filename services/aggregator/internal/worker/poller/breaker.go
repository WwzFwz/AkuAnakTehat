package poller

import "time"

// Breaker is owned by one endpoint fetcher. Only one probe can run because
// cycles never overlap; successful probes close the breaker.
type Breaker struct {
	Threshold int
	Cooldown  time.Duration
	failures  int
	until     time.Time
}

func (b *Breaker) Allow(now time.Time) bool { return !now.Before(b.until) }
func (b *Breaker) Success()                 { b.failures = 0; b.until = time.Time{} }
func (b *Breaker) Failure(now time.Time) {
	b.failures++
	if b.failures >= b.Threshold {
		b.until = now.Add(b.Cooldown)
	}
}
