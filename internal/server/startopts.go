package server

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func saveStartOpts(dataDir string, opts startOpts) {
	path := filepath.Join(dataDir, "start.json")
	data, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0o644)
}

func loadStartOpts(dataDir string) (startOpts, bool) {
	path := filepath.Join(dataDir, "start.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return startOpts{}, false
	}
	var opts startOpts
	if err := json.Unmarshal(data, &opts); err != nil {
		return startOpts{}, false
	}
	return opts, true
}
