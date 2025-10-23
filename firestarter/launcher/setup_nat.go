package launcher

func SetupNAT(ipAddr string) error {
	// Enable IP forwarding
	if err := RunCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// Set up iptables rules
	if err := RunCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ipAddr, "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "br0", "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", "br0", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}
