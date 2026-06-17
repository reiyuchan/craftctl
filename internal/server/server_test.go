package server

import (
	"testing"

	"github.com/reiyuchan/craftctl/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.New()
	s := New(cfg)
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.root == nil {
		t.Error("root fiber app is nil")
	}
	if s.mc == nil {
		t.Error("mc server is nil")
	}
	if s.ws == nil {
		t.Error("websocket is nil")
	}
}
