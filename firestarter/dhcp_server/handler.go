package dhcpserver

import (
	"log"
	"net"
	"strings"

	"github.com/devdevaraj/firestarter/init_app"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

func Handler(
	conn net.PacketConn,
	peer net.Addr,
	req *dhcpv4.DHCPv4,
	macToIP map[string]net.IP,
	cfg init_app.Config,
) (*dhcpv4.DHCPv4, error) {
	clientMAC := strings.ToLower(req.ClientHWAddr.String())
	ip, exists := macToIP[clientMAC]
	if !exists {
		log.Printf("Unknown MAC: %s", clientMAC)
		return nil, nil // Ignore/deny unknown MACs
	}
	resp, err := dhcpv4.NewReplyFromRequest(req)
	if err != nil {
		return nil, err
	}
	resp.YourIPAddr = ip

	gatewayIP, gatewayMask, err := ParseCIDR(*cfg.BridgeIP)
	if err != nil {
		log.Fatalf("Invalid CIDR: %v", err)
	}

	// Add other DHCP options as needed
	resp.Options.Update(dhcpv4.OptSubnetMask(gatewayMask))
	resp.Options.Update(dhcpv4.OptRouter(gatewayIP)) // Gateway
	resp.Options.Update(dhcpv4.OptDNS(
		net.ParseIP(cfg.Nameservers.NS1),
		net.ParseIP(cfg.Nameservers.NS2),
	)) // DNS

	log.Printf("Assigned %s to %s", ip.String(), clientMAC)
	return resp, nil
}
