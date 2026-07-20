package server

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

type WorldInfo struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	SizeBytes    int64  `json:"sizeBytes"`
	ModifiedDate string `json:"modifiedDate"`
	HasLevelDat  bool   `json:"hasLevelDat"`
	Active       bool   `json:"active"`
}

func (h Handler) listWorlds(c *fiber.Ctx) error {
	serverDir := h.cfg.ServerDir
	entries, err := os.ReadDir(serverDir)
	if err != nil {
		return errorResp(c, 500, fmt.Errorf("read server dir: %w", err))
	}

	props, _ := readServerProperties(serverDir)
	activeLevel := props.LevelName

	var worlds []WorldInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		worldPath := filepath.Join(serverDir, entry.Name())
		levelDat := filepath.Join(worldPath, "level.dat")
		info, err := os.Stat(levelDat)
		hasLevelDat := err == nil

		if !hasLevelDat {
			continue
		}

		size, _ := dirSize(worldPath)
		worlds = append(worlds, WorldInfo{
			Name:         entry.Name(),
			Size:         formatBytes(size),
			SizeBytes:    size,
			ModifiedDate: info.ModTime().Format(time.RFC3339),
			HasLevelDat:  hasLevelDat,
			Active:       entry.Name() == activeLevel,
		})
	}

	if worlds == nil {
		worlds = []WorldInfo{}
	}
	return c.JSON(worlds)
}

func (h Handler) loadWorld(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	if body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("world name is required"))
	}

	worldPath := filepath.Join(h.cfg.ServerDir, body.Name)
	if !existsFile(filepath.Join(worldPath, "level.dat")) {
		return errorResp(c, 404, fmt.Errorf("world not found: %s", body.Name))
	}

	props, err := readServerProperties(h.cfg.ServerDir)
	if err != nil {
		return errorResp(c, 500, fmt.Errorf("read server properties: %w", err))
	}

	props.LevelName = body.Name
	if err := writeServerProperties(h.cfg.ServerDir, props); err != nil {
		return errorResp(c, 500, fmt.Errorf("write server properties: %w", err))
	}

	return c.JSON(fiber.Map{"status": "ok", "active": body.Name})
}

func (h Handler) backupWorld(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	if body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("world name is required"))
	}

	worldPath := filepath.Join(h.cfg.ServerDir, body.Name)
	if !existsFile(filepath.Join(worldPath, "level.dat")) {
		return errorResp(c, 404, fmt.Errorf("world not found: %s", body.Name))
	}

	backupDir := filepath.Join(h.cfg.ServerDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return errorResp(c, 500, fmt.Errorf("create backup dir: %w", err))
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	zipName := fmt.Sprintf("%s_%s.zip", body.Name, timestamp)
	zipPath := filepath.Join(backupDir, zipName)

	if err := zipDir(worldPath, zipPath); err != nil {
		return errorResp(c, 500, fmt.Errorf("create backup: %w", err))
	}

	return c.JSON(fiber.Map{"status": "ok", "path": zipPath, "name": zipName})
}

func (h Handler) deleteWorld(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return errorResp(c, 400, fmt.Errorf("world name is required"))
	}

	if h.mc.IsRunning() {
		props, _ := readServerProperties(h.cfg.ServerDir)
		if props.LevelName == name {
			return errorResp(c, 400, fmt.Errorf("cannot delete active world while server is running"))
		}
	}

	worldPath := filepath.Join(h.cfg.ServerDir, name)
	if !existsFile(filepath.Join(worldPath, "level.dat")) {
		return errorResp(c, 404, fmt.Errorf("world not found: %s", name))
	}

	if err := os.RemoveAll(worldPath); err != nil {
		return errorResp(c, 500, fmt.Errorf("delete world: %w", err))
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

func zipDir(source, dest string) error {
	zipFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	return filepath.Walk(source, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		f, err := w.Create(rel)
		if err != nil {
			return fmt.Errorf("zip create entry: %w", err)
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer src.Close()

		if _, err := io.Copy(f, src); err != nil {
			return fmt.Errorf("zip write: %w", err)
		}
		return nil
	})
}


