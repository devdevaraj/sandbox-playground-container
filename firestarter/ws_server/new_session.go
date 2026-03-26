package wsserver

import (
	"fmt"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Session struct {
	ID         string
	sshClient  *ssh.Client
	sshSession *ssh.Session
	stdin      io.WriteCloser

	mu        sync.Mutex
	history   []byte
	listeners map[chan []byte]bool
}

type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]*Session
}

var globalSessions = &SessionManager{
	sessions: make(map[string]*Session),
}

const maxHistoryBytes = 1024 * 1024 // 1MB

func (sm *SessionManager) GetOrCreate(id string, config *ssh.ClientConfig, host, port string, cols, rows int) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, exists := sm.sessions[id]
	if exists {
		return s, nil
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("request for pseudo terminal failed: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	session.Stderr = session.Stdout // Redirect stderr to stdout

	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to start shell: %v", err)
	}

	s = &Session{
		ID:         id,
		sshClient:  client,
		sshSession: session,
		stdin:      stdin,
		history:    make([]byte, 0, maxHistoryBytes),
		listeners:  make(map[chan []byte]bool),
	}

	sm.sessions[id] = s

	// Start reading stdout
	go s.readLoop(stdout)

	// Clean up when SSH session ends
	go func() {
		session.Wait()
		log.Printf("Session %s ended by remote server", id)
		sm.mu.Lock()
		delete(sm.sessions, id)
		sm.mu.Unlock()

		s.mu.Lock()
		for ch := range s.listeners {
			close(ch)
		}
		s.listeners = make(map[chan []byte]bool)
		s.mu.Unlock()
	}()

	return s, nil
}

func (s *Session) readLoop(r io.Reader) {
	buf := make([]byte, 8192)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])

			s.mu.Lock()
			// Append to history, keeping max size
			s.history = append(s.history, chunk...)
			if len(s.history) > maxHistoryBytes {
				s.history = s.history[len(s.history)-maxHistoryBytes:]
			}

			// Broadcast to listeners
			for ch := range s.listeners {
				select {
				case ch <- chunk:
				default:
					log.Printf("Listener channel full for session %s, dropping some output", s.ID)
				}
			}
			s.mu.Unlock()
		}
		if err != nil {
			break
		}
	}
}

func (s *Session) AddListener() (chan []byte, []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan []byte, 100)
	s.listeners[ch] = true

	// Return current history
	hist := make([]byte, len(s.history))
	copy(hist, s.history)

	return ch, hist
}

func (s *Session) RemoveListener(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listeners[ch] {
		delete(s.listeners, ch)
		close(ch)
	}
}

func (s *Session) Write(data []byte) (int, error) {
	return s.stdin.Write(data)
}

func (s *Session) Resize(cols, rows int) error {
	return s.sshSession.WindowChange(rows, cols)
}
