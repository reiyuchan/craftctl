package mc

import (
	"io"
	"os/exec"
	"sync"
)

type Server struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	mu        sync.RWMutex
	listeners map[chan string]struct{}
	lmu       sync.RWMutex
	callbacks map[chan string]struct{}
	cbMu      sync.Mutex
	done      chan struct{}
}

func New() *Server {
	return &Server{
		listeners: make(map[chan string]struct{}),
		callbacks: make(map[chan string]struct{}),
	}
}

func (s *Server) Subscribe() chan string {
	ch := make(chan string, 256)
	s.lmu.Lock()
	s.listeners[ch] = struct{}{}
	s.lmu.Unlock()
	return ch
}

func (s *Server) Unsubscribe(ch chan string) {
	s.lmu.Lock()
	defer s.lmu.Unlock()
	if _, ok := s.listeners[ch]; ok {
		close(ch)
		delete(s.listeners, ch)
	}
}

func (s *Server) RegisterCallback(ch chan string) {
	s.cbMu.Lock()
	s.callbacks[ch] = struct{}{}
	s.cbMu.Unlock()
}

func (s *Server) UnregisterCallback(ch chan string) {
	s.cbMu.Lock()
	defer s.cbMu.Unlock()
	delete(s.callbacks, ch)
}

func (s *Server) broadcast(line string) {
	s.lmu.RLock()
	defer s.lmu.RUnlock()
	for ch := range s.listeners {
		select {
		case ch <- line:
		default:
		}
	}
	s.cbMu.Lock()
	for ch := range s.callbacks {
		select {
		case ch <- line:
		default:
		}
	}
	s.cbMu.Unlock()
}

func (s *Server) closeAllListeners() {
	s.lmu.Lock()
	defer s.lmu.Unlock()
	for ch := range s.listeners {
		close(ch)
	}
	s.listeners = make(map[chan string]struct{})
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) Start(java string, dir string, args ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		return nil
	}

	s.done = make(chan struct{})
	s.cmd = exec.Command(java, args...)
	s.cmd.Dir = dir
	s.cmd.Stdin = nil

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return err
	}

	s.stdin, err = s.cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := s.cmd.Start(); err != nil {
		return err
	}

	go s.readLoop(stdout, "OUT")
	go s.readLoop(stderr, "ERR")
	go s.awaitExit()

	return nil
}

func (s *Server) readLoop(r io.Reader, prefix string) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			s.broadcast(string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				s.broadcast(prefix + " ERROR: " + err.Error() + "\n")
			}
			break
		}
	}
}

func (s *Server) awaitExit() {
	s.mu.RLock()
	cmd := s.cmd
	done := s.done
	s.mu.RUnlock()
	if cmd == nil || done == nil {
		return
	}
	cmd.Wait()
	s.mu.Lock()
	isCurrent := s.cmd == cmd
	if isCurrent {
		s.cmd = nil
		s.stdin = nil
	}
	s.mu.Unlock()
	if isCurrent {
		s.closeAllListeners()
		close(done)
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	if s.cmd == nil || s.cmd.Process == nil {
		s.mu.Unlock()
		return nil
	}
	if s.stdin != nil {
		s.stdin.Write([]byte("stop\n"))
	}
	done := s.done
	s.mu.Unlock()
	<-done
	return nil
}

func (s *Server) Send(line string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.stdin == nil {
		return nil
	}

	_, err := s.stdin.Write([]byte(line + "\n"))
	return err
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.cmd != nil && s.cmd.Process != nil
}

func (s *Server) PID() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cmd == nil || s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}
