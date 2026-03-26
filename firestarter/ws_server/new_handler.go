package wsserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, host, port, VMID string, template init_app.Template, signer ssh.Signer) {
	vars := mux.Vars(r)
	sessionID := vars["session"]
	// sessionID := r.URL.Query().Get("sessionId")
	// if sessionID == "" {
	// 	http.Error(w, "missing sessionId", http.StatusBadRequest)
	// 	return
	// }

	cols, _ := strconv.Atoi(r.URL.Query().Get("cols"))
	rows, _ := strconv.Atoi(r.URL.Query().Get("rows"))
	perm := r.URL.Query().Get("perm")
	if cols == 0 {
		cols = 80
	}
	if rows == 0 {
		rows = 24
	}
	if perm == "" {
		perm = "rw"
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()

	config := &ssh.ClientConfig{
		User: *template.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	session, err := globalSessions.GetOrCreate(VMID+":"+sessionID, config, host, port, cols, rows)
	if err != nil {
		log.Printf("Failed to create/get SSH session: %v", err)
		conn.WriteMessage(websocket.BinaryMessage, []byte("\r\n\033[31mFailed to connect to SSH server: "+err.Error()+"\033[0m\r\n"))
		return
	}

	ch, history := session.AddListener()
	defer session.RemoveListener(ch)

	if len(history) > 0 {
		conn.WriteMessage(websocket.BinaryMessage, history)
	}

	// Write loop
	go func() {
		for data := range ch {
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
		// Session closed
		conn.WriteMessage(websocket.BinaryMessage, []byte("\r\n\033[31mSSH session ended.\033[0m\r\n"))
		conn.Close()
	}()

	// Read loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var pingMessage PingMessage
		if err := json.Unmarshal(msg, &pingMessage); err == nil && pingMessage.Type == "ping" {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`)); err != nil {
				log.Printf("Failed to resize PTY: %v", err)
			}
			continue
		}
		if perm == "rw" {
			session.Write(msg)
		}
	}
}

func handleResize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vm := vars["vm"]
	sessionID := vars["session"]

	cols, _ := strconv.Atoi(r.URL.Query().Get("cols"))
	rows, _ := strconv.Atoi(r.URL.Query().Get("rows"))
	if cols == 0 || rows == 0 {
		http.Error(w, "missing dimensions", http.StatusBadRequest)
		return
	}

	globalSessions.mu.Lock()
	session, ok := globalSessions.sessions[vm+":"+sessionID]
	globalSessions.mu.Unlock()

	if ok {
		session.Resize(cols, rows)
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "session not found", http.StatusNotFound)
	}
}
