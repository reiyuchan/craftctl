package server

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reiyuchan/craftctl/internal/mc"
)

const (
	maxDataPoints   = 60
	tpsProbeEvery   = 15 * time.Second
	maxTPSMisses    = 6
	tpsResponseWait = 5 * time.Second
)

type StatsSnapshot struct {
	CPU        float64 `json:"cpu"`
	RAM        int64   `json:"ram"`
	RAMPercent float64 `json:"ramPercent"`
	Threads    int     `json:"threads"`
	TPS        float64 `json:"tps"`
	Timestamp  int64   `json:"timestamp"`
}

type StatsCollector struct {
	mu       sync.RWMutex
	pid      int
	mc       *mc.Server
	tps      float64
	data     []StatsSnapshot
	stopCh   chan struct{}
	running  bool
	onUpdate func(StatsSnapshot)
}

func NewStatsCollector(mcSrv *mc.Server) *StatsCollector {
	return &StatsCollector{
		mc:   mcSrv,
		data: make([]StatsSnapshot, 0, maxDataPoints),
	}
}

func (sc *StatsCollector) SetOnUpdate(fn func(StatsSnapshot)) {
	sc.mu.Lock()
	sc.onUpdate = fn
	sc.mu.Unlock()
}

func (sc *StatsCollector) Start(pid int) {
	sc.mu.Lock()
	if sc.running {
		sc.mu.Unlock()
		return
	}
	sc.pid = pid
	sc.data = sc.data[:0]
	sc.tps = 0
	sc.stopCh = make(chan struct{})
	sc.running = true
	sc.mu.Unlock()
	go sc.loop()
	go sc.probeLoop()
}

func (sc *StatsCollector) Stop() {
	sc.mu.Lock()
	if !sc.running {
		sc.mu.Unlock()
		return
	}
	sc.running = false
	sc.mu.Unlock()
	close(sc.stopCh)
}

func (sc *StatsCollector) loop() {
	sc.mu.RLock()
	stopCh := sc.stopCh
	sc.mu.RUnlock()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			snap := sc.collect()
			sc.mu.Lock()
			sc.data = append(sc.data, snap)
			if len(sc.data) > maxDataPoints {
				sc.data = sc.data[len(sc.data)-maxDataPoints:]
			}
			fn := sc.onUpdate
			sc.mu.Unlock()
			if fn != nil {
				fn(snap)
			}
		case <-stopCh:
			return
		}
	}
}

func (sc *StatsCollector) collect() StatsSnapshot {
	sc.mu.RLock()
	pid := sc.pid
	tps := sc.tps
	sc.mu.RUnlock()

	now := time.Now().UnixMilli()
	snap := StatsSnapshot{Timestamp: now}

	switch runtime.GOOS {
	case "linux":
		snap = sc.collectLinux(pid)
	case "darwin":
		snap = sc.collectDarwin(pid)
	default:
		snap = sc.collectFallback(pid)
	}
	snap.TPS = tps
	snap.Timestamp = now
	return snap
}

func (sc *StatsCollector) collectLinux(pid int) StatsSnapshot {
	snap := StatsSnapshot{Timestamp: time.Now().UnixMilli()}

	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return snap
	}

	contents := string(data)
	lParen := strings.Index(contents, "(")
	rParen := strings.LastIndex(contents, ")")
	if lParen < 0 || rParen < 0 {
		return snap
	}

	fields := strings.Fields(contents[rParen+2:])
	if len(fields) >= 20 {
		numThreads, _ := strconv.ParseInt(fields[17], 10, 64)
		snap.Threads = int(numThreads)
	}

	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusData, err := os.ReadFile(statusPath)
	if err == nil {
		sc := bufio.NewScanner(strings.NewReader(string(statusData)))
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "VmRSS:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					kb, _ := strconv.ParseInt(parts[1], 10, 64)
					snap.RAM = kb * 1024
				}
			}
		}
	}

	if len(fields) >= 20 {
		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		starttime, _ := strconv.ParseUint(fields[19], 10, 64)

		clkTck := uint64(100)

		uptimeBytes, err := os.ReadFile("/proc/uptime")
		if err == nil {
			uptimeStr := strings.Fields(string(uptimeBytes))[0]
			uptime, _ := strconv.ParseFloat(uptimeStr, 64)

			totalTicks := utime + stime
			secondsSinceStart := (uptime * float64(clkTck)) - float64(starttime)
			if secondsSinceStart > 0 {
				snap.CPU = (float64(totalTicks) / secondsSinceStart) * 100.0
			}
		}
	}

	if snap.RAM > 0 {
		memInfo, err := os.ReadFile("/proc/meminfo")
		if err == nil {
			sc := bufio.NewScanner(strings.NewReader(string(memInfo)))
			for sc.Scan() {
				line := sc.Text()
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						totalKB, _ := strconv.ParseInt(parts[1], 10, 64)
						if totalKB > 0 {
							snap.RAMPercent = float64(snap.RAM) / float64(totalKB*1024) * 100.0
						}
					}
					break
				}
			}
		}
	}

	return snap
}

func (sc *StatsCollector) collectDarwin(pid int) StatsSnapshot {
	snap := StatsSnapshot{Timestamp: time.Now().UnixMilli()}

	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu,rss,nlwp", "-comm").Output()
	if err != nil {
		return snap
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return snap
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 3 {
		snap.CPU, _ = strconv.ParseFloat(fields[0], 64)
		rssKB, _ := strconv.ParseInt(fields[1], 10, 64)
		snap.RAM = rssKB * 1024
		snap.Threads, _ = strconv.Atoi(fields[2])
	}

	totalMem, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err == nil {
		total, _ := strconv.ParseInt(strings.TrimSpace(string(totalMem)), 10, 64)
		if total > 0 {
			snap.RAMPercent = float64(snap.RAM) / float64(total) * 100.0
		}
	}

	return snap
}

func (sc *StatsCollector) collectFallback(pid int) StatsSnapshot {
	snap := StatsSnapshot{Timestamp: time.Now().UnixMilli()}

	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu,rss").Output()
	if err != nil {
		return snap
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return snap
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 2 {
		snap.CPU, _ = strconv.ParseFloat(fields[0], 64)
		rssKB, _ := strconv.ParseInt(fields[1], 10, 64)
		snap.RAM = rssKB * 1024
	}

	return snap
}

func (sc *StatsCollector) Snapshot() StatsSnapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if len(sc.data) == 0 {
		return StatsSnapshot{Timestamp: time.Now().UnixMilli()}
	}
	return sc.data[len(sc.data)-1]
}

func (sc *StatsCollector) History() []StatsSnapshot {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	out := make([]StatsSnapshot, len(sc.data))
	copy(out, sc.data)
	return out
}

// ── TPS probe ────────────────────────────────────────────────────────────────

var tpsLineRe = regexp.MustCompile(`TPS from last 1m, 5m, 15m:\s*([\d.]+),\s*([\d.]+),\s*([\d.]+)`)

func parseTPSLine(line string) (float64, bool) {
	m := tpsLineRe.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	tps, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return tps, true
}

func (sc *StatsCollector) probeLoop() {
	sc.mu.RLock()
	stopCh := sc.stopCh
	sc.mu.RUnlock()
	ticker := time.NewTicker(tpsProbeEvery)
	defer ticker.Stop()
	misses := 0
	for {
		select {
		case <-ticker.C:
			matched, stopped := sc.probe(stopCh)
			if stopped {
				return
			}
			if matched {
				misses = 0
				continue
			}
			misses++
			if misses >= maxTPSMisses {
				return
			}
		case <-stopCh:
			return
		}
	}
}

func (sc *StatsCollector) probe(stopCh chan struct{}) (matched, stopped bool) {
	sc.mu.RLock()
	mcSrv := sc.mc
	sc.mu.RUnlock()
	if mcSrv == nil || !mcSrv.IsRunning() {
		return false, false
	}

	ch := make(chan string, 8)
	mcSrv.RegisterCallback(ch)
	defer mcSrv.UnregisterCallback(ch)

	if err := mcSrv.Send("tps"); err != nil {
		return false, false
	}

	timeout := time.After(tpsResponseWait)
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return false, false
			}
			if tps, ok := parseTPSLine(line); ok {
				sc.mu.Lock()
				sc.tps = tps
				sc.mu.Unlock()
				return true, false
			}
		case <-stopCh:
			return false, true
		case <-timeout:
			return false, false
		}
	}
}
