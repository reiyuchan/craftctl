package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ScheduledTask struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Interval string `json:"interval"`
	Enabled  bool   `json:"enabled"`
	LastRun  string `json:"lastRun"`
	NextRun  string `json:"nextRun"`
}

type Scheduler struct {
	tasks   []ScheduledTask
	mu      sync.Mutex
	dataDir string
	mc      mcServerOps
	ws      *WebSocket
	logger  *zap.Logger
	stopCh  chan struct{}
}

type mcServerOps interface {
	IsRunning() bool
	Start(java string, dir string, args ...string) error
	Stop() error
}

func NewScheduler(dataDir string, mc mcServerOps, ws *WebSocket, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		dataDir: dataDir,
		mc:      mc,
		ws:      ws,
		logger:  logger,
		stopCh:  make(chan struct{}),
	}
}

func (s *Scheduler) load() {
	path := filepath.Join(s.dataDir, "scheduler.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var tasks []ScheduledTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return
	}
	s.tasks = tasks
	s.recalcNextRuns()
}

func (s *Scheduler) save() {
	path := filepath.Join(s.dataDir, "scheduler.json")
	os.MkdirAll(filepath.Dir(path), 0o755)
	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0o644)
}

func (s *Scheduler) recalcNextRuns() {
	now := time.Now()
	for i := range s.tasks {
		if !s.tasks[i].Enabled || s.tasks[i].Interval == "" {
			continue
		}
		dur, err := parseInterval(s.tasks[i].Interval)
		if err != nil {
			continue
		}
		last, _ := time.Parse(time.RFC3339, s.tasks[i].LastRun)
		if last.IsZero() {
			s.tasks[i].NextRun = now.Add(dur).Format(time.RFC3339)
		} else {
			next := last.Add(dur)
			if next.Before(now) {
				next = now.Add(dur)
			}
			s.tasks[i].NextRun = next.Format(time.RFC3339)
		}
	}
}

func (s *Scheduler) Start() {
	s.load()
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case now := <-ticker.C:
			s.tick(now)
		}
	}
}

func (s *Scheduler) tick(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		t := &s.tasks[i]
		if !t.Enabled || t.Interval == "" {
			continue
		}
		next, err := time.Parse(time.RFC3339, t.NextRun)
		if err != nil {
			continue
		}
		if now.After(next) || now.Equal(next) {
			s.executeTask(t)
			dur, err := parseInterval(t.Interval)
			if err != nil {
				continue
			}
			t.LastRun = now.Format(time.RFC3339)
			t.NextRun = now.Add(dur).Format(time.RFC3339)
		}
	}
	s.save()
}

func (s *Scheduler) executeTask(t *ScheduledTask) {
	s.logger.Info("executing scheduled task", zap.String("name", t.Name), zap.String("type", t.Type))
	switch t.Type {
	case "backup":
		s.runBackup()
	case "restart":
		s.runRestart()
	case "stop":
		s.runStop()
	}
}

func (s *Scheduler) runBackup() {
	if s.mc == nil {
		return
	}
	serverDir := filepath.Join(s.dataDir, "servers", "default")
	props, err := readServerProperties(serverDir)
	if err != nil {
		s.logger.Error("scheduler backup: read props", zap.Error(err))
		return
	}
	worldName := props.LevelName
	if worldName == "" {
		worldName = "world"
	}
	worldPath := filepath.Join(serverDir, worldName)
	if !existsFile(filepath.Join(worldPath, "level.dat")) {
		s.logger.Error("scheduler backup: world not found", zap.String("world", worldName))
		return
	}
	backupDir := filepath.Join(serverDir, "backups")
	os.MkdirAll(backupDir, 0o755)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	zipName := fmt.Sprintf("%s_%s.zip", worldName, timestamp)
	zipPath := filepath.Join(backupDir, zipName)
	if err := zipDir(worldPath, zipPath); err != nil {
		s.logger.Error("scheduler backup: create zip", zap.Error(err))
		return
	}
	s.logger.Info("scheduled backup completed", zap.String("path", zipPath))
}

func (s *Scheduler) runRestart() {
	if s.mc == nil {
		return
	}
	if s.mc.IsRunning() {
		if err := s.mc.Stop(); err != nil {
			s.logger.Error("scheduler restart: stop failed", zap.Error(err))
			return
		}
	}
	serverDir := filepath.Join(s.dataDir, "servers", "default")
	java := "java"
	args := []string{"-Xms2G", "-Xmx4G", "-jar", "server.jar", "nogui"}
	eulaPath := filepath.Join(serverDir, "eula.txt")
	if !existsFile(eulaPath) {
		os.WriteFile(eulaPath, []byte("eula=true\n"), 0o644)
	}
	if err := s.mc.Start(java, serverDir, args...); err != nil {
		s.logger.Error("scheduler restart: start failed", zap.Error(err))
		return
	}
	s.logger.Info("scheduled restart completed")
}

func (s *Scheduler) runStop() {
	if s.mc == nil {
		return
	}
	if err := s.mc.Stop(); err != nil {
		s.logger.Error("scheduler stop failed", zap.Error(err))
		return
	}
	s.logger.Info("scheduled stop completed")
}

func (s *Scheduler) GetTasks() []ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ScheduledTask, len(s.tasks))
	copy(out, s.tasks)
	return out
}

func (s *Scheduler) AddTask(task ScheduledTask) ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.ID = fmt.Sprintf("task_%d", time.Now().UnixMilli())
	s.tasks = append(s.tasks, task)
	s.recalcNextRuns()
	s.save()
	for i := range s.tasks {
		if s.tasks[i].ID == task.ID {
			return s.tasks[i]
		}
	}
	return task
}

func (s *Scheduler) UpdateTask(id string, update ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Name = update.Name
			s.tasks[i].Type = update.Type
			s.tasks[i].Interval = update.Interval
			s.tasks[i].Enabled = update.Enabled
			s.recalcNextRuns()
			s.save()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func (s *Scheduler) RemoveTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			s.save()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func (s *Scheduler) EnableTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Enabled = true
			s.recalcNextRuns()
			s.save()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func (s *Scheduler) DisableTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Enabled = false
			s.save()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func (s *Scheduler) RunTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.executeTask(&s.tasks[i])
			s.tasks[i].LastRun = time.Now().Format(time.RFC3339)
			s.recalcNextRuns()
			s.save()
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func parseInterval(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty interval")
	}
	var num int
	var unit string
	for i, ch := range s {
		if ch >= '0' && ch <= '9' {
			continue
		}
		num = 0
		fmt.Sscanf(s[:i], "%d", &num)
		unit = s[i:]
		break
	}
	if num == 0 {
		fmt.Sscanf(s, "%d", &num)
	}
	if num <= 0 {
		return 0, fmt.Errorf("invalid interval: %s", s)
	}
	switch unit {
	case "m", "min", "minute", "minutes":
		return time.Duration(num) * time.Minute, nil
	case "h", "hr", "hour", "hours":
		return time.Duration(num) * time.Hour, nil
	case "d", "day", "days":
		return time.Duration(num) * 24 * time.Hour, nil
	case "w", "week", "weeks":
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}
