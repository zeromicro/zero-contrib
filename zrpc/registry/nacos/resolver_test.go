package nacos

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
)

type testNamingClient struct {
	mu         sync.Mutex
	closeCalls int
}

func (c *testNamingClient) CloseClient() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeCalls++
}

func (c *testNamingClient) CloseCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closeCalls
}

func TestResolverCloseClosesNamingClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &testNamingClient{}
	r := &resolvr{
		cancelFunc: cancel,
		client:     client,
	}

	r.Close()
	r.Close()

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("resolver context was not canceled")
	}

	if got := client.CloseCalls(); got != 1 {
		t.Fatalf("CloseClient called %d times, want 1", got)
	}
}

func TestWatcherCallbackReturnsWhenResolverIsClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	watcher := newWatcher(ctx, cancel, make(chan []string))
	done := make(chan struct{})

	go func() {
		watcher.CallBackHandle([]model.Instance{{Ip: "127.0.0.1", Port: 8080}}, nil)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher callback did not return after resolver shutdown")
	}
}
