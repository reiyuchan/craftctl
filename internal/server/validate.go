package server

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
)

func (h Handler) validateStartup(opts startOpts) error {
	if !existsFile(filePath(h.cfg.ServerDir, "server.jar")) {
		return errNoServerJar
	}

	java := opts.JavaPath
	if java == "" {
		java = "java"
	}
	if strings.ContainsRune(java, filepath.Separator) {
		if !existsFile(java) {
			return fmt.Errorf("java executable not found: %s. Install Java or set a valid Java Path.", java)
		}
	} else {
		if _, err := exec.LookPath(java); err != nil {
			return fmt.Errorf("java executable not found: %s. Install Java or set a valid Java Path.", java)
		}
	}

	props, err := readServerProperties(h.cfg.ServerDir)
	if err != nil {
		return fmt.Errorf("read server properties: %w", err)
	}
	if props.ServerPort > 0 {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", props.ServerPort))
		if err != nil {
			return fmt.Errorf("port %d is already in use", props.ServerPort)
		}
		ln.Close()
	}

	return nil
}
