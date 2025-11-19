package launcher

import "log"

func SetupNAT(ipAddr string, bridgeName string) error {
	log.Printf("Setting NAT for Bridge %s with ip %s", bridgeName, ipAddr)
	// Enable IP forwarding
	if err := RunCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// Set up iptables rules
	if err := RunCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ipAddr, "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}
