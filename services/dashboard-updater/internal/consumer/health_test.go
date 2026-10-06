package consumer

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func TestPingDeadlineWithUnresponsiveBroker(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() { defer conn.Close(); <-done }()
		}
	}()
	client, err := kgo.NewClient(kgo.SeedBrokers(listener.Addr().String()))
	if err != nil {
		listener.Close()
		close(done)
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close(); close(done); client.Close() })
	c := &Consumer{Client: client}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	err = c.Ping(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Ping error = %v; want deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("Ping ignored caller deadline: %v", elapsed)
	}
	// The first underlying handshake is still blocked. Further calls must
	// fail promptly rather than launching more probes against that broker.
	for i := 0; i < 10; i++ {
		next, stop := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := c.Ping(next)
		stop()
		if err == nil || errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("concurrent probe was not rejected promptly: %v", err)
		}
	}

}
