package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const maxLogReadBytes int64 = 2 << 20 // 2MB

type LogFileInfo struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	SizeBytes    int64  `json:"sizeBytes"`
	ModifiedDate string `json:"modifiedDate"`
	IsGzipped    bool   `json:"isGzipped"`
}

func (h Handler) listLogs(c *fiber.Ctx) error {
	logsDir := filepath.Join(h.cfg.ServerDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return c.JSON([]LogFileInfo{})
		}
		return errorResp(c, 500, fmt.Errorf("read logs dir: %w", err))
	}

	result := make([]LogFileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, LogFileInfo{
			Name:         e.Name(),
			Size:         formatBytes(info.Size()),
			SizeBytes:    info.Size(),
			ModifiedDate: info.ModTime().Format(time.RFC3339),
			IsGzipped:    strings.HasSuffix(e.Name(), ".gz"),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return c.JSON(result)
}

func (h Handler) resolveLogPath(name string) (string, error) {
	name = safeName(name)
	if name == "" {
		return "", fmt.Errorf("log file name is required")
	}
	logsDir := filepath.Join(h.cfg.ServerDir, "logs")
	abs := filepath.Join(logsDir, name)
	rel, err := filepath.Rel(logsDir, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || filepath.HasPrefix(rel, "..") {
		return "", fmt.Errorf("access denied")
	}
	return abs, nil
}

func (h Handler) readLog(c *fiber.Ctx) error {
	path, err := h.resolveLogPath(c.Query("file"))
	if err != nil {
		return errorResp(c, 400, err)
	}
	if !existsFile(path) {
		return errorResp(c, 404, fmt.Errorf("log file not found"))
	}

	content, truncated, err := readLogTail(path, maxLogReadBytes)
	if err != nil {
		return errorResp(c, 500, err)
	}

	return c.JSON(fiber.Map{
		"name":      filepath.Base(path),
		"content":   content,
		"truncated": truncated,
	})
}

func (h Handler) downloadLog(c *fiber.Ctx) error {
	path, err := h.resolveLogPath(c.Query("file"))
	if err != nil {
		return errorResp(c, 400, err)
	}
	if !existsFile(path) {
		return errorResp(c, 404, fmt.Errorf("log file not found"))
	}
	return c.Download(path, filepath.Base(path))
}

func (h Handler) deleteLog(c *fiber.Ctx) error {
	name := safeName(c.Params("name"))
	path, err := h.resolveLogPath(name)
	if err != nil {
		return errorResp(c, 400, err)
	}
	if !existsFile(path) {
		return errorResp(c, 404, fmt.Errorf("log file not found"))
	}
	if err := os.Remove(path); err != nil {
		return errorResp(c, 500, fmt.Errorf("delete log: %w", err))
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func readLogTail(path string, maxBytes int64) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", false, fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return "", false, fmt.Errorf("open gzip: %w", err)
		}
		defer gz.Close()
		r = gz
	}

	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return "", false, fmt.Errorf("read log: %w", err)
	}

	truncated := int64(len(data)) > maxBytes
	if truncated {
		data = data[:maxBytes]
	}
	return string(data), truncated, nil
}
