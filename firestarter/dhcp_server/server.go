package dhcpserver

// import (
// 	"log"
// 	"net"

// 	"github.com/devdevaraj/firestarter/init_app"
// 	"github.com/insomniacslk/dhcp/dhcpv4"
// 	"github.com/insomniacslk/dhcp/dhcpv4/server4"
// )

// var macToIP = map[string]net.IP{
// 	"aa:fc:00:00:00:01": net.IPv4(172, 16, 0, 2),
// 	"fc:fc:00:00:01:01": net.IPv4(172, 16, 0, 3),
// }

// func HandlerFunc(
// 	macToIP map[string]net.IP,
// 	cfg init_app.Config,
// ) server4.Handler {
// 	return func(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
// 		Handler(conn, peer, req, macToIP, cfg)
// 		// return resp, error
// 	}
// }

// func DHCPServer(cfg init_app.Config) {
// 	iface := *cfg.Bridge
// 	localAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 67}

// 	server, err := server4.NewServer(
// 		iface,
// 		localAddr,
// 		HandlerFunc(macToIP, cfg),
// 	)

// 	if err != nil {
// 		log.Fatalf("Failed to create server: %v", err)
// 	}
// 	log.Fatal(server.Serve())
// }
