package test_temp

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

// PacketType defines the type of message exchanged.
type PacketType int

const (
	PacketHello PacketType = iota
	PacketPing
	PacketData
	PacketResize
	PacketExit
	PacketError
)

type Packet struct {
	Type PacketType  `json:"type"`
	ID   string      `json:"id"`
	Args interface{} `json:"args,omitempty"`
}

// Define argument structs for each packet type as needed.
type HelloArgs struct {
	Token string `json:"token"`
	Cols  int    `json:"cols"`
	Rows  int    `json:"rows"`
	Seq   int    `json:"seq"`
}
type DataArgs struct {
	Data string `json:"data"`
}
type ResizeArgs struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}
type ErrorArgs struct {
	Reason string `json:"reason"`
}

type TerminalSession struct {
	id         string
	pty        *os.File
	cmd        *exec.Cmd
	clients    map[*websocket.Conn]bool
	buffer     []string
	bufferLock sync.Mutex
	sequence   int
	mux        sync.Mutex
}

type SessionManager struct {
	sessions map[string]*TerminalSession
	mux      sync.Mutex
}

var manager = &SessionManager{
	sessions: make(map[string]*TerminalSession),
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Read first message to get session ID
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	var pkt Packet
	json.Unmarshal(msg, &pkt)
	sessionID := pkt.ID

	// Get or create session
	session := manager.GetOrCreateSession(sessionID)
	session.AddClient(conn)

	// Start goroutines for reading from client and writing PTY output
	go session.ReadFromPTY()
	go session.WriteToPTY(conn)

	// Main loop: handle incoming messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var pkt Packet
		json.Unmarshal(msg, &pkt)
		session.HandleMessage(conn, pkt)
	}
	session.RemoveClient(conn)
}
