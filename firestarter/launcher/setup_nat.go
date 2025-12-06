package launcher

import "log"

func SetupNAT(networkAddr string, bridgeName string) error {
	log.Printf("Setting NAT for Bridge %s with ip %s", bridgeName, networkAddr)
	// Enable IP forwarding
	if err := RunCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	// Set up iptables rules
	if err := RunCommand("iptables", "-F", "FORWARD"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-t", "nat", "-F", "POSTROUTING"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", networkAddr, "-o", "eth0", "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-s", "192.168.0.0/16", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-s", "10.0.0.0/8", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-s", "172.16.0.0/12", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-d", "10.0.0.0/8", "-j", "DROP"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-d", "172.16.0.0/12", "-j", "DROP"); err != nil {
		return err
	}
	if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-d", "192.168.0.0/16", "-j", "DROP"); err != nil {
		return err
	}

	if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
		return err
	}

	// if err := RunCommand("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", ipAddr, "-o", "eth0", "-j", "MASQUERADE"); err != nil {
	// 	return err
	// }
	// if err := RunCommand("iptables", "-A", "FORWARD", "-i", bridgeName, "-o", "eth0", "-j", "ACCEPT"); err != nil {
	// 	return err
	// }
	// if err := RunCommand("iptables", "-A", "FORWARD", "-i", "eth0", "-o", bridgeName, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
	// 	return err
	// }

	return nil
}
