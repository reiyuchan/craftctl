package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/pflag"
)

type Config struct {
	Port      int
	ServerDir string
	DataDir   string
	MaxRAM    string
	MinRAM    string
	JVMFlags  string
	JavaPath  string
}

const (
	Port     = 8000
	MaxRAM   = "4G"
	MinRAM   = "2G"
	JVMFlags = "-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40"
)

var (
	CustomPort   int
	registerOnce sync.Once
)

func New() Config {
	registerOnce.Do(func() {
		pflag.IntVar(&CustomPort, "port", 0, "sets the server startup port")
		pflag.Parse()
	})

	port := CustomPort
	if port == 0 {
		port = Port
	}

	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "ctlcraft")

	return Config{
		Port:      port,
		DataDir:   dataDir,
		ServerDir: filepath.Join(dataDir, "servers", "default"),
		MaxRAM:    MaxRAM,
		MinRAM:    MinRAM,
		JVMFlags:  JVMFlags,
	}
}
