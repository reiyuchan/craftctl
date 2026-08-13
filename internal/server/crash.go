package server

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CrashWatcher struct {
	srv     *Server
	logger  *zap.Logger
	stopCh  chan struct{}
	stopped bool

	mu         sync.Mutex
	manualStop bool
	opts       startOpts
}

func NewCrashWatcher(s *Server, logger *zap.Logger) *CrashWatcher {
	return &CrashWatcher{srv: s, logger: logger, stopCh: make(chan struct{})}
}

func (cw *CrashWatcher) Start() {
	go cw.run()
}

func (cw *CrashWatcher) Stop() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.stopped {
		return
	}
	cw.stopped = true
	close(cw.stopCh)
}

func (cw *CrashWatcher) markManualStop() {
	cw.mu.Lock()
	cw.manualStop = true
	cw.mu.Unlock()
}

func (cw *CrashWatcher) recordStart(opts startOpts) {
	cw.mu.Lock()
	cw.opts = opts
	cw.mu.Unlock()
}

func (cw *CrashWatcher) isManualStop() bool {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.manualStop {
		cw.manualStop = false
		return true
	}
	return false
}

func (cw *CrashWatcher) run() {
	for {
		select {
		case <-cw.stopCh:
			return
		default:
		}
		ch := cw.srv.mc.Subscribe()
		for range ch {
		}
		if cw.isManualStop() {
			continue
		}
		cw.autoRestart()
	}
}

func (cw *CrashWatcher) autoRestart() {
	cfg := loadCrashConfig(cw.srv.cfg.DataDir)
	if !cfg.Enabled {
		msg := "Server crashed unexpectedly. Auto-restart is disabled."
		cw.srv.NotifyWebhook("crash", msg)
		appendHistory(cw.srv.cfg.DataDir, "crash", msg)
		cw.logger.Warn("server crashed; auto-restart disabled")
		return
	}

	cooldown := time.Duration(cfg.CooldownSeconds) * time.Second
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		cw.mu.Lock()
		opts := cw.opts
		cw.mu.Unlock()

		msg := fmt.Sprintf("Server crashed unexpectedly (attempt %d/%d). Restarting in %ds.", attempt, cfg.MaxRetries, cfg.CooldownSeconds)
		cw.srv.NotifyWebhook("crash", msg)
		appendHistory(cw.srv.cfg.DataDir, "crash", msg)
		cw.logger.Warn("server crashed", zap.Int("attempt", attempt), zap.Int("max", cfg.MaxRetries))

		select {
		case <-time.After(cooldown):
		case <-cw.stopCh:
			return
		}

		if cw.srv.mc.IsRunning() {
			cw.logger.Info("server already running; skipping crash auto-restart")
			return
		}

		ch := cw.srv.mc.Subscribe()
		if err := cw.restart(opts); err != nil {
			cw.srv.mc.Unsubscribe(ch)
			cw.logger.Error("crash auto-restart failed", zap.Error(err))
			return
		}
		for range ch {
		}
		if cw.isManualStop() {
			return
		}
	}
	giveUpMsg := fmt.Sprintf("Server crashed %d times in a row; giving up.", cfg.MaxRetries)
	cw.srv.NotifyWebhook("crash", giveUpMsg)
	appendHistory(cw.srv.cfg.DataDir, "crash", giveUpMsg)
	cw.logger.Error("server crashed repeatedly; giving up")
}

func (cw *CrashWatcher) restart(opts startOpts) error {
	if opts.JavaPath == "" {
		opts.JavaPath = "java"
	}
	if opts.MinRAM == "" {
		opts.MinRAM = cw.srv.cfg.MinRAM
	}
	if opts.MaxRAM == "" {
		opts.MaxRAM = cw.srv.cfg.MaxRAM
	}
	if opts.JVMFlags == "" {
		opts.JVMFlags = cw.srv.cfg.JVMFlags
	}

	if err := cw.srv.ws.Start(opts.JavaPath, cw.srv.cfg.ServerDir, serverArgs(opts)...); err != nil {
		return err
	}
	cw.srv.stats.Stop()
	if pid := cw.srv.mc.PID(); pid > 0 {
		cw.srv.stats.Start(pid)
	}
	cw.logger.Info("server auto-restarted after crash")
	return nil
}

func serverArgs(opts startOpts) []string {
	args := []string{"-Xms" + opts.MinRAM, "-Xmx" + opts.MaxRAM}
	args = append(args, fields(opts.JVMFlags)...)
	return append(args, "-jar", "server.jar", "nogui")
}
