package server

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BackupInfo struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	SizeBytes    int64  `json:"sizeBytes"`
	ModifiedDate string `json:"modifiedDate"`
	Type         string `json:"type"`
}

func (h Handler) listBackups(c *fiber.Ctx) error {
	backupDir := filepath.Join(h.cfg.ServerDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return c.JSON([]BackupInfo{})
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:         entry.Name(),
			Size:         formatBytes(info.Size()),
			SizeBytes:    info.Size(),
			ModifiedDate: info.ModTime().Format(time.RFC3339),
			Type:         backupType(entry.Name()),
		})
	}

	if backups == nil {
		backups = []BackupInfo{}
	}
	return c.JSON(backups)
}

func (h Handler) createFullBackup(c *fiber.Ctx) error {
	if h.mc.IsRunning() {
		return errorResp(c, 400, fmt.Errorf("cannot create backup while server is running"))
	}

	backupDir := filepath.Join(h.cfg.ServerDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return errorResp(c, 500, fmt.Errorf("create backup dir: %w", err))
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	zipName := fmt.Sprintf("full_%s.zip", timestamp)
	zipPath := filepath.Join(backupDir, zipName)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return errorResp(c, 500, fmt.Errorf("create zip: %w", err))
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	dirsToBackup := []string{"world", "world_nether", "world_the_end"}
	filesToBackup := []string{"server.properties"}

	for _, dir := range dirsToBackup {
		dirPath := filepath.Join(h.cfg.ServerDir, dir)
		if existsDir(dirPath) {
			if err := addDirToZip(w, h.cfg.ServerDir, dirPath); err != nil {
				return errorResp(c, 500, fmt.Errorf("zip %s: %w", dir, err))
			}
		}
	}

	for _, f := range filesToBackup {
		fPath := filepath.Join(h.cfg.ServerDir, f)
		if existsFile(fPath) {
			addFileToZip(w, h.cfg.ServerDir, fPath)
		}
	}

	for _, dir := range []string{"mods", "plugins"} {
		dirPath := filepath.Join(h.cfg.ServerDir, dir)
		if existsDir(dirPath) {
			addDirToZip(w, h.cfg.ServerDir, dirPath)
		}
	}

	return c.JSON(fiber.Map{"status": "ok", "path": zipPath, "name": zipName})
}

func (h Handler) restoreBackup(c *fiber.Ctx) error {
	var body struct{ Name string }
	if err := c.BodyParser(&body); err != nil {
		return errorResp(c, 400, err)
	}
	if body.Name == "" {
		return errorResp(c, 400, fmt.Errorf("backup name is required"))
	}

	if h.mc.IsRunning() {
		return errorResp(c, 400, fmt.Errorf("cannot restore backup while server is running"))
	}

	backupPath := filepath.Join(h.cfg.ServerDir, "backups", body.Name)
	if !existsFile(backupPath) {
		return errorResp(c, 404, fmt.Errorf("backup not found: %s", body.Name))
	}

	r, err := zip.OpenReader(backupPath)
	if err != nil {
		return errorResp(c, 500, fmt.Errorf("open backup: %w", err))
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(h.cfg.ServerDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return errorResp(c, 500, fmt.Errorf("create dir: %w", err))
		}
		outFile, err := os.Create(fpath)
		if err != nil {
			return errorResp(c, 500, fmt.Errorf("create file: %w", err))
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return errorResp(c, 500, fmt.Errorf("open zip entry: %w", err))
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return errorResp(c, 500, fmt.Errorf("extract file: %w", err))
		}
	}

	return c.JSON(fiber.Map{"status": "restored", "name": body.Name})
}

func (h Handler) deleteBackup(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return errorResp(c, 400, fmt.Errorf("backup name is required"))
	}

	backupPath := filepath.Join(h.cfg.ServerDir, "backups", name)
	if !existsFile(backupPath) {
		return errorResp(c, 404, fmt.Errorf("backup not found: %s", name))
	}

	if err := os.Remove(backupPath); err != nil {
		return errorResp(c, 500, fmt.Errorf("delete backup: %w", err))
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

func backupType(name string) string {
	if strings.HasPrefix(name, "full_") {
		return "full"
	}
	return "world"
}

func existsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func addDirToZip(w *zip.Writer, baseDir, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		addFileToZip(w, baseDir, path)
		return nil
	})
}

func addFileToZip(w *zip.Writer, baseDir, filePath string) {
	rel, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return
	}
	rel = filepath.ToSlash(rel)

	f, err := w.Create(rel)
	if err != nil {
		return
	}

	src, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer src.Close()

	io.Copy(f, src)
}
