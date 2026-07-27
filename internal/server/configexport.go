package server

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h Handler) exportConfig(c *fiber.Ctx) error {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	files := []string{"server.properties", "eula.txt", "scheduler.json"}
	for _, f := range files {
		path := filepath.Join(h.cfg.ServerDir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		w, err := zipWriter.Create(f)
		if err != nil {
			continue
		}
		w.Write(data)
	}
	if err := zipWriter.Close(); err != nil {
		return errorResp(c, 500, err)
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", `attachment; filename="craftctl-config.zip"`)
	return c.Send(buf.Bytes())
}

func (h Handler) importConfig(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return errorResp(c, 400, fmt.Errorf("no file provided: %w", err))
	}
	src, err := file.Open()
	if err != nil {
		return errorResp(c, 500, err)
	}
	defer src.Close()

	zipReader, err := zip.NewReader(src, file.Size)
	if err != nil {
		return errorResp(c, 400, fmt.Errorf("invalid zip: %w", err))
	}

	var extracted []string
	for _, f := range zipReader.File {
		name := filepath.Base(f.Name)
		if strings.Contains(f.Name, "..") {
			continue
		}
		dst := filepath.Join(h.cfg.ServerDir, name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(dst, 0o755)
			continue
		}
		os.MkdirAll(filepath.Dir(dst), 0o755)
		out, err := os.Create(dst)
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			continue
		}
		io.Copy(out, rc)
		rc.Close()
		out.Close()
		extracted = append(extracted, name)
	}
	return c.JSON(fiber.Map{"status": "ok", "files": extracted})
}
