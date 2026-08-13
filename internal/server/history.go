package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HistoryEvent struct {
	Event   string `json:"event"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

const maxHistoryEvents = 100

var historyMu sync.Mutex

func appendHistory(dataDir, event, message string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	path := filepath.Join(dataDir, "history.json")
	events := readHistoryFile(path)
	events = append(events, HistoryEvent{Event: event, Time: time.Now().Format(time.RFC3339), Message: message})
	if len(events) > maxHistoryEvents {
		events = events[len(events)-maxHistoryEvents:]
	}
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0o644)
}

func readHistory(dataDir string) []HistoryEvent {
	historyMu.Lock()
	defer historyMu.Unlock()
	return readHistoryFile(filepath.Join(dataDir, "history.json"))
}

func readHistoryFile(path string) []HistoryEvent {
	data, err := os.ReadFile(path)
	if err != nil {
		return []HistoryEvent{}
	}
	var events []HistoryEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return []HistoryEvent{}
	}
	if events == nil {
		events = []HistoryEvent{}
	}
	return events
}
