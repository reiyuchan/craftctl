package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

var listCmdRe = regexp.MustCompile(`^There are (\d+)/\d+ players online:\s*(.*)$`)

func (h Handler) registerPlayerRoutes(g fiber.Router) {
	g.Get("/players", h.listPlayers)
	g.Get("/players/ops", h.getOps)
	g.Post("/players/op", h.opPlayer)
	g.Post("/players/deop", h.deopPlayer)
	g.Post("/players/kick", h.kickPlayer)
	g.Post("/players/ban", h.banPlayer)
	g.Post("/players/pardon", h.pardonPlayer)
	g.Get("/players/whitelist", h.getWhitelist)
	g.Post("/players/whitelist/add", h.whitelistAdd)
	g.Post("/players/whitelist/remove", h.whitelistRemove)
}

func (h Handler) requireRunning(c *fiber.Ctx) error {
	if !h.mc.IsRunning() {
		return errorResp(c, 503, fmt.Errorf("server not running"))
	}
	return nil
}

func (h Handler) listPlayers(c *fiber.Ctx) error {
	if err := h.requireRunning(c); err != nil {
		return err
	}

	ch := make(chan string, 8)
	h.mc.RegisterCallback(ch)
	defer h.mc.UnregisterCallback(ch)

	if err := h.mc.Send("list"); err != nil {
		return errorResp(c, 500, fmt.Errorf("send list command: %w", err))
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return errorResp(c, 500, fmt.Errorf("channel closed"))
			}
			m := listCmdRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			total := 0
			fmt.Sscanf(m[1], "%d", &total)
			var names []string
			if trimmed := strings.TrimSpace(m[2]); trimmed != "" {
				names = strings.Split(trimmed, ", ")
			}
			return c.JSON(fiber.Map{
				"total":   total,
				"players": names,
			})
		case <-timeout:
			return errorResp(c, 504, fmt.Errorf("timeout waiting for list response"))
		}
	}
}

func (h Handler) getOps(c *fiber.Ctx) error {
	ops, err := readJSONList(h.cfg.ServerDir, "ops.json")
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(ops)
}

func (h Handler) opPlayer(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("op " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) deopPlayer(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("deop " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) kickPlayer(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("kick " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) banPlayer(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("ban " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) pardonPlayer(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("pardon " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) getWhitelist(c *fiber.Ctx) error {
	list, err := readJSONList(h.cfg.ServerDir, "whitelist.json")
	if err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(list)
}

func (h Handler) whitelistAdd(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("whitelist add " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h Handler) whitelistRemove(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("name required"))
	}
	if err := h.requireRunning(c); err != nil {
		return err
	}
	if err := h.mc.Send("whitelist remove " + body.Name); err != nil {
		return errorResp(c, 500, err)
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func readJSONList(serverDir, filename string) ([]map[string]interface{}, error) {
	path := filepath.Join(serverDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	return list, nil
}
