package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCrashConfigDefaults(t *testing.T) {
	cfg := loadCrashConfig(t.TempDir())
	if !cfg.Enabled {
		t.Errorf("Enabled = %v, want true", cfg.Enabled)
	}
	if cfg.CooldownSeconds != 30 {
		t.Errorf("CooldownSeconds = %d, want 30", cfg.CooldownSeconds)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
}

func TestCrashConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := CrashConfig{Enabled: false, CooldownSeconds: 60, MaxRetries: 5}
	if err := saveCrashConfig(dir, cfg); err != nil {
		t.Fatalf("saveCrashConfig: %v", err)
	}
	got := loadCrashConfig(dir)
	if got.Enabled != cfg.Enabled || got.CooldownSeconds != cfg.CooldownSeconds || got.MaxRetries != cfg.MaxRetries {
		t.Errorf("loadCrashConfig = %+v, want %+v", got, cfg)
	}
}

func TestCrashConfigPartiallyFilledFillsDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "crash.json"), []byte(`{"enabled":false}`), 0o644); err != nil {
		t.Fatalf("write crash.json: %v", err)
	}
	cfg := loadCrashConfig(dir)
	if cfg.Enabled {
		t.Error("Enabled = true, want false")
	}
	if cfg.CooldownSeconds != 30 {
		t.Errorf("CooldownSeconds = %d, want 30 (default filled)", cfg.CooldownSeconds)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3 (default filled)", cfg.MaxRetries)
	}
}
