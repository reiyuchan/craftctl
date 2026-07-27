package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type WebhookConfig struct {
	URL     string   `json:"url"`
	Enabled bool     `json:"enabled"`
	Events  []string `json:"events"`
}

func webhookPath(dataDir string) string {
	return filepath.Join(dataDir, "webhook.json")
}

func loadWebhookConfig(dataDir string) WebhookConfig {
	var cfg WebhookConfig
	data, err := os.ReadFile(webhookPath(dataDir))
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveWebhookConfig(dataDir string, cfg WebhookConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(webhookPath(dataDir), data, 0o644)
}

func sendWebhook(url, content string) error {
	if url == "" {
		return fmt.Errorf("webhook URL not configured")
	}
	_, err := httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"content": content}).
		Post(url)
	return err
}

func (h Handler) getWebhookConfig(c *fiber.Ctx) error {
	cfg := loadWebhookConfig(h.cfg.DataDir)
	return c.JSON(cfg)
}

func (h Handler) updateWebhookConfig(c *fiber.Ctx) error {
	var cfg WebhookConfig
	if err := c.BodyParser(&cfg); err != nil {
		return errorResp(c, 400, err)
	}
	if err := saveWebhookConfig(h.cfg.DataDir, cfg); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h Handler) testWebhook(c *fiber.Ctx) error {
	cfg := loadWebhookConfig(h.cfg.DataDir)
	if cfg.URL == "" {
		return errorResp(c, 400, fmt.Errorf("webhook URL not configured"))
	}
	if err := sendWebhook(cfg.URL, "CraftCTL test notification"); err != nil {
		return errorResp(c, 500, fmt.Errorf("webhook failed: %w", err))
	}
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Server) NotifyWebhook(event, message string) {
	cfg := loadWebhookConfig(h.cfg.DataDir)
	if !cfg.Enabled || cfg.URL == "" {
		return
	}
	for _, e := range cfg.Events {
		if e == event || e == "*" {
			sendWebhook(cfg.URL, fmt.Sprintf("**[%s]** %s", event, message))
			return
		}
	}
}
