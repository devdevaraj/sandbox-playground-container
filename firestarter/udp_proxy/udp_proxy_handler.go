package udpproxy

import (
	"log"
	"net"
	"time"
)

type UDPProxy_t struct {
	TargetAddr *net.UDPAddr
}

func NewUDPProxy(target string) *UDPProxy_t {
	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		log.Fatalf("Failed to resolve UDP target %s: %v", target, err)
	}
	return &UDPProxy_t{TargetAddr: addr}
}

func (p *UDPProxy_t) Start(listenAddr string) {
	laddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP listen address %s: %v", listenAddr, err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalf("Failed to start UDP listener on %s: %v", listenAddr, err)
	}
	defer conn.Close()

	log.Printf("UDP proxy listening on %s → %s\n", listenAddr, p.TargetAddr)

	buffer := make([]byte, 64*1024) // 64KB buffer

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("UDP read error:", err)
			continue
		}

		go p.handlePacket(conn, clientAddr, buffer[:n])
	}
}

func (p *UDPProxy_t) handlePacket(conn *net.UDPConn, clientAddr *net.UDPAddr, data []byte) {
	backendConn, err := net.DialUDP("udp", nil, p.TargetAddr)
	if err != nil {
		log.Println("Error dialing backend:", err)
		return
	}
	defer backendConn.Close()

	// Send packet to backend
	_, err = backendConn.Write(data)
	if err != nil {
		log.Println("Error writing to backend:", err)
		return
	}

	// Wait for response
	backendConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply := make([]byte, 64*1024)
	n, _, err := backendConn.ReadFromUDP(reply)
	if err != nil {
		return
	}

	// Send back to client
	_, err = conn.WriteToUDP(reply[:n], clientAddr)
	if err != nil {
		log.Println("Error writing back to client:", err)
	}
}
