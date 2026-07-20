package mc

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.listeners == nil {
		t.Error("listeners map is nil")
	}
}

func TestServerIsRunning(t *testing.T) {
	s := New()
	if s.IsRunning() {
		t.Error("IsRunning() = true, want false")
	}
}

func TestServerSubscribe(t *testing.T) {
	s := New()
	ch := s.Subscribe()
	if ch == nil {
		t.Error("Subscribe() returned nil")
	}
	s.Unsubscribe(ch)
}

func TestServerSendBeforeStart(t *testing.T) {
	s := New()
	err := s.Send("test")
	if err != nil {
		t.Errorf("Send() before start: %v", err)
	}
}

func TestServerStopBeforeStart(t *testing.T) {
	s := New()
	err := s.Stop()
	if err != nil {
		t.Errorf("Stop() before start: %v", err)
	}
}
