package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogRetentionSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := LogRetentionConfig{KeepDays: 14}
	if err := saveLogRetention(dir, cfg); err != nil {
		t.Fatalf("saveLogRetention: %v", err)
	}
	got := loadLogRetention(dir)
	if got.KeepDays != 14 {
		t.Errorf("loadLogRetention keepDays = %d, want 14", got.KeepDays)
	}
}

func TestLogRetentionLoadMissingDefault(t *testing.T) {
	got := loadLogRetention(t.TempDir())
	if got.KeepDays != 0 {
		t.Errorf("loadLogRetention default keepDays = %d, want 0", got.KeepDays)
	}
}

func TestPruneOldLogsDeletesOldKeepsRecent(t *testing.T) {
	logsDir := t.TempDir()
	recent := filepath.Join(logsDir, "latest.log")
	old1 := filepath.Join(logsDir, "2024-01-01-1.log.gz")
	old2 := filepath.Join(logsDir, "2024-01-02-2.log.gz")
	for _, p := range []string{recent, old1, old2} {
		if err := os.WriteFile(p, []byte("log"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	oldTime := time.Now().AddDate(0, 0, -30)
	for _, p := range []string{old1, old2} {
		if err := os.Chtimes(p, oldTime, oldTime); err != nil {
			t.Fatalf("chtimes %s: %v", p, err)
		}
	}
	os.MkdirAll(filepath.Join(logsDir, "subdir"), 0o755)

	deleted, err := pruneOldLogs(logsDir, 7)
	if err != nil {
		t.Fatalf("pruneOldLogs: %v", err)
	}
	if deleted != 2 {
		t.Errorf("pruned %d, want 2", deleted)
	}
	if _, err := os.Stat(recent); err != nil {
		t.Errorf("recent file deleted: %v", err)
	}
	for _, p := range []string{old1, old2} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("old file %s still exists", p)
		}
	}
}

func TestPruneOldLogsDisabled(t *testing.T) {
	logsDir := t.TempDir()
	os.WriteFile(filepath.Join(logsDir, "latest.log"), []byte("log"), 0o644)
	deleted, err := pruneOldLogs(logsDir, 0)
	if err != nil {
		t.Fatalf("pruneOldLogs(0): %v", err)
	}
	if deleted != 0 {
		t.Errorf("pruned %d, want 0", deleted)
	}
}

func TestPruneOldLogsMissingDir(t *testing.T) {
	deleted, err := pruneOldLogs(filepath.Join(t.TempDir(), "nope"), 7)
	if err != nil {
		t.Fatalf("pruneOldLogs missing dir: %v", err)
	}
	if deleted != 0 {
		t.Errorf("pruned %d, want 0", deleted)
	}
}
