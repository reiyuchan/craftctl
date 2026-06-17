package main

import (
	"testing"

	"github.com/reiyuchan/craftctl/internal/config"
	"github.com/reiyuchan/craftctl/internal/server"
)

func TestRunWiring(t *testing.T) {
	cfg := config.Config{
		Port:      config.Port,
		ServerDir: t.TempDir(),
		DataDir:   t.TempDir(),
		MaxRAM:    config.MaxRAM,
		MinRAM:    config.MinRAM,
		JVMFlags:  config.JVMFlags,
	}
	s := server.New(cfg)
	if s == nil {
		t.Fatal("server.New returned nil")
	}
}
