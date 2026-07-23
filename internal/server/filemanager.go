package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

type fileEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

func (h Handler) resolvePath(rel string) (string, error) {
	if rel == "" {
		return h.cfg.ServerDir, nil
	}
	abs := filepath.Join(h.cfg.ServerDir, filepath.Clean("/"+rel))
	rel, err := filepath.Rel(h.cfg.ServerDir, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || filepath.HasPrefix(rel, "..") {
		return "", fmt.Errorf("access denied")
	}
	return abs, nil
}

func (h Handler) listFiles(c *fiber.Ctx) error {
	rel := c.Query("path", "")
	abs, err := h.resolvePath(rel)
	if err != nil {
		return errorResp(c, 403, err)
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON([]fileEntry{})
		}
		return errorResp(c, 500, fmt.Errorf("read directory: %w", err))
	}

	result := make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, fileEntry{
			Name:    e.Name(),
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format(time.RFC3339),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})

	return c.JSON(result)
}

func (h Handler) readFile(c *fiber.Ctx) error {
	rel := c.Query("path", "")
	if rel == "" {
		return errorResp(c, 400, fmt.Errorf("path is required"))
	}

	abs, err := h.resolvePath(rel)
	if err != nil {
		return errorResp(c, 403, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return errorResp(c, 404, fmt.Errorf("file not found"))
	}
	if info.IsDir() {
		return errorResp(c, 400, fmt.Errorf("cannot read a directory"))
	}
	if info.Size() > 1<<20 {
		return errorResp(c, 400, fmt.Errorf("file too large to read (max 1MB)"))
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return errorResp(c, 500, fmt.Errorf("read file: %w", err))
	}

	return c.JSON(fiber.Map{
		"path":    rel,
		"content": string(data),
		"size":    info.Size(),
	})
}

func (h Handler) writeFile(c *fiber.Ctx) error {
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	if body.Path == "" {
		return errorResp(c, 400, fmt.Errorf("path is required"))
	}

	abs, err := h.resolvePath(body.Path)
	if err != nil {
		return errorResp(c, 403, err)
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return errorResp(c, 500, fmt.Errorf("create directory: %w", err))
	}
	if err := os.WriteFile(abs, []byte(body.Content), 0o644); err != nil {
		return errorResp(c, 500, fmt.Errorf("write file: %w", err))
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func (h Handler) deleteFile(c *fiber.Ctx) error {
	rel := c.Query("path", "")
	if rel == "" {
		return errorResp(c, 400, fmt.Errorf("path is required"))
	}

	abs, err := h.resolvePath(rel)
	if err != nil {
		return errorResp(c, 403, err)
	}

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return errorResp(c, 404, fmt.Errorf("file not found"))
	}

	if err := os.RemoveAll(abs); err != nil {
		return errorResp(c, 500, fmt.Errorf("delete: %w", err))
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h Handler) makeDir(c *fiber.Ctx) error {
	var body struct{ Path string }
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	if body.Path == "" {
		return errorResp(c, 400, fmt.Errorf("path is required"))
	}

	abs, err := h.resolvePath(body.Path)
	if err != nil {
		return errorResp(c, 403, err)
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		return errorResp(c, 500, fmt.Errorf("create directory: %w", err))
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
