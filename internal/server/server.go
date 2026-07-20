package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/reiyuchan/craftctl/internal/config"
	"github.com/reiyuchan/craftctl/internal/mc"
	"github.com/reiyuchan/craftctl/internal/ui"
	"go.uber.org/zap"
)

type Server struct {
	root   *fiber.App
	cfg    config.Config
	logger *zap.Logger
	mc     *mc.Server
	ws     *WebSocket
	events    *EventHub
	stats     *StatsCollector
}

func New(cfg config.Config) *Server {
	app := fiber.New()
	l, _ := zap.NewProduction()
	app.Use(recover.New())
	app.Use(logger.New())

	app.Use("/", ui.Handler())

	mcServer := mc.New()
	events := NewEventHub(mcServer, l)
	ws := NewWebSocket(l, mcServer)
	stats := NewStatsCollector()
	stats.SetOnUpdate(func(snap StatsSnapshot) {
		events.BroadcastStats(snap)
	})

	s := &Server{root: app, cfg: cfg, logger: l, mc: mcServer, ws: ws, events: events, stats: stats}

	h := newHandler(s)
	h.routes(app)
	app.Use("/ws", ws.Handler())

	return s
}

func (s *Server) Listen() error {
	defer s.logger.Sync()
	s.logger.Info(fmt.Sprintf("server listening on port %d", s.cfg.Port))
	return s.root.Listen(fmt.Sprintf(":%d", s.cfg.Port))
}

func (s *Server) Stop() error {
	s.logger.Info("Shutting down...")
	if err := s.ws.Stop(); err != nil {
		s.logger.Error("Failed to stop MC server", zap.Error(err))
	}
	return s.root.Shutdown()
}
