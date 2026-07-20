package server

import (
	"bufio"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/reiyuchan/craftctl/internal/mc"
	"go.uber.org/zap"
)

type EventHub struct {
	mc     *mc.Server
	logger *zap.Logger

	mu          sync.RWMutex
	log         map[chan string]struct{}
	stopped     map[chan string]struct{}
	errors      map[chan string]struct{}
	stats       map[chan string]struct{}
	stoppedFlag bool
}

func NewEventHub(mc *mc.Server, logger *zap.Logger) *EventHub {
	h := &EventHub{
		mc:      mc,
		logger:  logger,
		log:     make(map[chan string]struct{}),
		stopped: make(map[chan string]struct{}),
		errors:  make(map[chan string]struct{}),
		stats:   make(map[chan string]struct{}),
	}
	go h.run()
	return h
}

func (h *EventHub) run() {
	for {
		h.mu.Lock()
		h.stoppedFlag = false
		h.mu.Unlock()

		ch := h.mc.Subscribe()
		for line := range ch {
			h.broadcast(h.log, line)
			if strings.Contains(line, "ERROR") {
				h.broadcast(h.errors, line)
			}
		}
		h.mu.Lock()
		h.stoppedFlag = true
		h.mu.Unlock()
		h.broadcast(h.stopped, "")
	}
}

func (h *EventHub) broadcast(clients map[chan string]struct{}, data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range clients {
		select {
		case ch <- data:
		default:
		}
	}
}

func (h *EventHub) BroadcastStats(snap StatsSnapshot) {
	data := fmt.Sprintf(`{"cpu":%.2f,"ram":%d,"ramPercent":%.2f,"threads":%d,"timestamp":%d}`,
		snap.CPU, snap.RAM, snap.RAMPercent, snap.Threads, snap.Timestamp)
	h.broadcast(h.stats, data)
}

func (h *EventHub) Handler(eventType string) fiber.Handler {
	var clients map[chan string]struct{}
	switch eventType {
	case "server-log":
		clients = h.log
	case "server-stopped":
		clients = h.stopped
	case "server-error":
		clients = h.errors
	case "server-stats":
		clients = h.stats
	default:
		return func(c *fiber.Ctx) error {
			return c.Status(404).JSON(fiber.Map{"error": "unknown event type"})
		}
	}

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		ch := make(chan string, 64)
		h.mu.Lock()
		clients[ch] = struct{}{}
		stopped := h.stoppedFlag
		h.mu.Unlock()

		if eventType == "server-stopped" && stopped {
			ch <- ""
		}

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			defer func() {
				h.mu.Lock()
				delete(clients, ch)
				h.mu.Unlock()
			}()

			for data := range ch {
				fmt.Fprintf(w, "data: %s\n\n", data)
				if err := w.Flush(); err != nil {
					return
				}
			}
		})

		return nil
	}
}
