package server

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type CrashConfig struct {
	Enabled         bool `json:"enabled"`
	CooldownSeconds int  `json:"cooldownSeconds"`
	MaxRetries      int  `json:"maxRetries"`
}

func defaultCrashConfig() CrashConfig {
	return CrashConfig{Enabled: true, CooldownSeconds: 30, MaxRetries: 3}
}

func crashConfigPath(dataDir string) string {
	return filepath.Join(dataDir, "crash.json")
}

func loadCrashConfig(dataDir string) CrashConfig {
	cfg := defaultCrashConfig()
	data, err := os.ReadFile(crashConfigPath(dataDir))
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveCrashConfig(dataDir string, cfg CrashConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(crashConfigPath(dataDir), data, 0o644)
}

func (h Handler) getCrashConfig(c *fiber.Ctx) error {
	return c.JSON(loadCrashConfig(h.cfg.DataDir))
}

func (h Handler) updateCrashConfig(c *fiber.Ctx) error {
	var cfg CrashConfig
	if err := c.BodyParser(&cfg); err != nil {
		return errorResp(c, 400, err)
	}
	if err := saveCrashConfig(h.cfg.DataDir, cfg); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
