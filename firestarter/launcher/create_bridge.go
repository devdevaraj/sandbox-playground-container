package launcher

func CreateBridge(ipAddr string) error {
	// Delete existing bridge if it exists
	RunCommand("ip", "link", "set", "br0", "down")
	RunCommand("ip", "link", "delete", "br0", "type", "bridge")

	// Create bridge
	if err := RunCommand("ip", "link", "add", "name", "br0", "type", "bridge"); err != nil {
		return err
	}

	// Assign IP address to bridge
	if err := RunCommand("ip", "addr", "add", ipAddr, "dev", "br0"); err != nil {
		return err
	}

	// Bring up the bridge interface
	if err := RunCommand("ip", "link", "set", "br0", "up"); err != nil {
		return err
	}

	return nil
}
