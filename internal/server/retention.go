package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogRetentionConfig struct {
	KeepDays int `json:"keepDays"`
}

func retentionPath(dataDir string) string {
	return filepath.Join(dataDir, "retention.json")
}

func loadLogRetention(dataDir string) LogRetentionConfig {
	var cfg LogRetentionConfig
	data, err := os.ReadFile(retentionPath(dataDir))
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveLogRetention(dataDir string, c LogRetentionConfig) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(retentionPath(dataDir), data, 0o644)
}

func (h Handler) getLogRetention(c *fiber.Ctx) error {
	cfg := loadLogRetention(h.cfg.DataDir)
	return c.JSON(fiber.Map{"keepDays": cfg.KeepDays})
}

func (h Handler) updateLogRetention(c *fiber.Ctx) error {
	var cfg LogRetentionConfig
	if err := c.BodyParser(&cfg); err != nil {
		return errorResp(c, 400, err)
	}
	if cfg.KeepDays < 0 {
		return errorResp(c, 400, fmt.Errorf("keepDays must be >= 0"))
	}
	if err := saveLogRetention(h.cfg.DataDir, cfg); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h Handler) pruneLogsNow(c *fiber.Ctx) error {
	cfg := loadLogRetention(h.cfg.DataDir)
	deleted, err := pruneOldLogs(filepath.Join(h.cfg.ServerDir, "logs"), cfg.KeepDays)
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"deleted": deleted})
}

func pruneOldLogs(logsDir string, keepDays int) (int, error) {
	if keepDays <= 0 {
		return 0, nil
	}
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read logs dir: %w", err)
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	deleted := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		if err := os.Remove(filepath.Join(logsDir, e.Name())); err != nil {
			continue
		}
		deleted++
	}
	return deleted, nil
}

func (h Handler) registerRetentionRoutes(g fiber.Router) {
	g.Get("/logs/retention", h.getLogRetention)
	g.Put("/logs/retention", h.updateLogRetention)
	g.Post("/logs/prune", h.pruneLogsNow)
}
