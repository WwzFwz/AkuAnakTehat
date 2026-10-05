package middleware

import "testing"

func TestLimitsDistinguishConcurrencyAndRateRejection(t *testing.T) {
	limits := New(100, 100, 1)
	release, reason := limits.Enter("client")
	if reason != Allowed || release == nil {
		t.Fatalf("first request rejected: release=%v reason=%d", release != nil, reason)
	}
	if second, reason := limits.Enter("client"); second != nil || reason != ConcurrencyLimited {
		t.Fatalf("second request release=%v reason=%d; want concurrency rejection", second != nil, reason)
	}
	release()

	rateLimited := New(1, 1, 2)
	first, reason := rateLimited.Enter("client")
	if reason != Allowed || first == nil {
		t.Fatalf("rate test first request rejected: reason=%d", reason)
	}
	first()
	if second, reason := rateLimited.Enter("client"); second != nil || reason != RateLimited {
		t.Fatalf("second request release=%v reason=%d; want rate rejection", second != nil, reason)
	}
}
