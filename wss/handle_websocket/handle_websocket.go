package handle_websocket

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type wsSession struct {
	client     *ssh.Client
	session    *ssh.Session
	stdin      io.WriteCloser
	stdout     io.Reader
	stderr     io.Reader
	created    time.Time
	lastActive time.Time
}

var (
	sessions      = make(map[string]*wsSession)
	sessionsMutex = sync.Mutex{}
)

type PingMessage struct {
	Type string `json:"type"`
}

type ResizeMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow any origin for demo purposes
	},
}

func fetchOrCreateSession(sessionID, ip string, config *ssh.ClientConfig) (*wsSession, error) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	// if sess := sessions[sessionID]; sess != nil {
	// 	sess.lastActive = time.Now()
	// 	log.Printf("Reusing SSH session for %s", sessionID)
	// 	return sess, nil
	// }

	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		return nil, fmt.Errorf("SSH dial error: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request pty: %w", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("error starting shell: %w", err)
	}

	wsSess := &wsSession{
		client:     client,
		session:    session,
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		created:    time.Now(),
		lastActive: time.Now(),
	}
	sessions[sessionID] = wsSess
	log.Printf("Created new SSH session for %s", sessionID)
	return wsSess, nil
}

func HandleWebsocket(w http.ResponseWriter, r *http.Request, ip string, VMID string) {
	vars := mux.Vars(r)
	sessionid := vars["session"]
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// Read SSH private key
	privateKeyBytes, err := os.ReadFile("/root/firecracker/keys/ubuntu-24.04.id_rsa")
	if err != nil {
		log.Printf("Failed to read private key: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to read private key: %v", err)))
		return
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		log.Printf("Failed to parse private key: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to parse private key: %v", err)))
		return
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	wsSess, err := fetchOrCreateSession(sessionid, ip, config)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	// Handle messages from WebSocket to SSH
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				wsSess.session.Close()
				return
			}

			var pingMessage PingMessage
			if err := json.Unmarshal(msg, &pingMessage); err == nil && pingMessage.Type == "ping" {
				if err := ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`)); err != nil {
					log.Printf("Failed to resize PTY: %v", err)
				}
				continue
			}

			var resizeMsg ResizeMessage
			if err := json.Unmarshal(msg, &resizeMsg); err == nil && resizeMsg.Type == "resize" {
				// Adjust PTY size
				if err := wsSess.session.WindowChange(resizeMsg.Rows, resizeMsg.Cols); err != nil {
					log.Printf("Failed to resize PTY: %v", err)
				}
				continue
			}

			if _, err := wsSess.stdin.Write(msg); err != nil {
				log.Printf("SSH write error: %v", err)
				return
			}
		}
	}()

	// Handle stdout from SSH to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := wsSess.stdout.Read(buffer)
			if err != nil {
				log.Printf("SSH stdout read error: %v", err)
				ws.Close()
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	// Handle stderr from SSH to WebSocket
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := wsSess.stderr.Read(buffer)
			if err != nil {
				log.Printf("SSH stderr read error: %v", err)
				return
			}
			message := fmt.Sprintf("STDERR: %s", buffer[:n])
			if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	<-r.Context().Done()
	log.Printf("WebSocket closed for session %s", sessionid)
	// Wait for SSH session to end
	if err := wsSess.session.Wait(); err != nil {
		log.Printf("SSH session ended with: %v", err)
	}
}
