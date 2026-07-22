package server

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h Handler) getServerProperties(c *fiber.Ctx) error {
	path := filepath.Join(h.cfg.ServerDir, "server.properties")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON(fiber.Map{})
		}
		return errorResp(c, 500, err)
	}

	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key != "" {
			props[key] = val
		}
	}

	return c.JSON(ToProps(props))
}

func (h Handler) updateServerProperties(c *fiber.Ctx) error {
	var props map[string]string
	if err := c.BodyParser(&props); err != nil {
		return errorResp(c, 400, err)
	}

	os.MkdirAll(h.cfg.ServerDir, 0o755)
	path := filepath.Join(h.cfg.ServerDir, "server.properties")

	var sb strings.Builder
	sb.WriteString("#Minecraft server properties\n")
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(props[k])
		sb.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return errorResp(c, 500, fmt.Errorf("write server.properties: %w", err))
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func ToProps(props map[string]string) fiber.Map {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(fiber.Map, len(props))
	for _, k := range keys {
		result[k] = props[k]
	}
	return result
}
