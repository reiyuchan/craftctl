package server

import (
	"sync"
	"testing"
	"time"

	"github.com/reiyuchan/craftctl/internal/mc"
	"go.uber.org/zap"
)

func TestEventHub_Creation(t *testing.T) {
	s := mc.New()
	logger := zap.NewNop()
	hub := NewEventHub(s, logger)
	if hub == nil {
		t.Fatal("NewEventHub returned nil")
	}
	if hub.log == nil {
		t.Error("log map is nil")
	}
	if hub.stopped == nil {
		t.Error("stopped map is nil")
	}
	if hub.errors == nil {
		t.Error("errors map is nil")
	}
}

func TestEventHub_Broadcast(t *testing.T) {
	s := mc.New()
	logger := zap.NewNop()
	hub := NewEventHub(s, logger)

	ch1 := make(chan string, 64)
	ch2 := make(chan string, 64)

	hub.mu.Lock()
	hub.log[ch1] = struct{}{}
	hub.log[ch2] = struct{}{}
	hub.mu.Unlock()

	hub.broadcast(hub.log, "test line")

	select {
	case msg := <-ch1:
		if msg != "test line" {
			t.Errorf("ch1 got %q, want %q", msg, "test line")
		}
	case <-time.After(time.Second):
		t.Error("ch1 timed out")
	}

	select {
	case msg := <-ch2:
		if msg != "test line" {
			t.Errorf("ch2 got %q, want %q", msg, "test line")
		}
	case <-time.After(time.Second):
		t.Error("ch2 timed out")
	}
}

func TestEventHub_BroadcastErrors(t *testing.T) {
	s := mc.New()
	logger := zap.NewNop()
	hub := NewEventHub(s, logger)

	ch := make(chan string, 64)
	hub.mu.Lock()
	hub.errors[ch] = struct{}{}
	hub.mu.Unlock()

	hub.broadcast(hub.errors, "ERROR something")

	select {
	case msg := <-ch:
		if msg != "ERROR something" {
			t.Errorf("got %q, want %q", msg, "ERROR something")
		}
	case <-time.After(time.Second):
		t.Error("timed out")
	}
}

func TestEventHub_Unsubscribe(t *testing.T) {
	s := mc.New()
	logger := zap.NewNop()
	hub := NewEventHub(s, logger)

	ch := make(chan string, 64)
	hub.mu.Lock()
	hub.log[ch] = struct{}{}
	hub.mu.Unlock()

	hub.mu.Lock()
	delete(hub.log, ch)
	hub.mu.Unlock()

	hub.broadcast(hub.log, "should not arrive")

	select {
	case <-ch:
		t.Error("received message after unsubscribe")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEventHub_BroadcastNonBlocking(t *testing.T) {
	s := mc.New()
	logger := zap.NewNop()
	hub := NewEventHub(s, logger)

	full := make(chan string)
	hub.mu.Lock()
	hub.log[full] = struct{}{}
	hub.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.broadcast(hub.log, "overflow")
		}()
	}
	wg.Wait()
}
