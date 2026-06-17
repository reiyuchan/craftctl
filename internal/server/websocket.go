package server

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/reiyuchan/craftctl/internal/mc"
	"go.uber.org/zap"
)

type WebSocket struct {
	logger *zap.Logger
	server *mc.Server
	mu     sync.RWMutex
	conns  map[*websocket.Conn]struct{}
}

func NewWebSocket(logger *zap.Logger, srv *mc.Server) *WebSocket {
	return &WebSocket{
		logger: logger,
		server: srv,
		conns:  make(map[*websocket.Conn]struct{}),
	}
}

func (ws *WebSocket) Handler() fiber.Handler {
	return websocket.New(ws.handle)
}

func (ws *WebSocket) handle(c *websocket.Conn) {
	ws.mu.Lock()
	ws.conns[c] = struct{}{}
	ws.mu.Unlock()

	defer func() {
		ws.mu.Lock()
		delete(ws.conns, c)
		ws.mu.Unlock()
		c.Close()
	}()

	go ws.streamToClient(c)

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				ws.logger.Info("websocket: client disconnected")
			} else {
				ws.logger.Error("websocket: read error", zap.Error(err))
			}
			break
		}

		if len(msg) == 0 {
			continue
		}

		if err := ws.server.Send(string(msg)); err != nil {
			ws.logger.Error("websocket: send error", zap.Error(err))
		}
	}
}

func (ws *WebSocket) streamToClient(c *websocket.Conn) {
	for line := range ws.server.Output() {
		if err := c.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}
}

func (ws *WebSocket) Start(java string, dir string, args ...string) error {
	return ws.server.Start(java, dir, args...)
}

func (ws *WebSocket) Stop() error {
	return ws.server.Stop()
}

func (ws *WebSocket) Send(line string) error {
	return ws.server.Send(line)
}

func (ws *WebSocket) IsRunning() bool {
	return ws.server.IsRunning()
}
