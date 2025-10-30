package launcher

func CreateBridge(ipAddr string, name string) error {
	// Delete existing bridge if it exists
	RunCommand("ip", "link", "set", name, "down")
	RunCommand("ip", "link", "delete", name, "type", "bridge")

	// Create bridge
	if err := RunCommand("ip", "link", "add", "name", name, "type", "bridge"); err != nil {
		return err
	}

	// Assign IP address to bridge
	if err := RunCommand("ip", "addr", "add", ipAddr, "dev", name); err != nil {
		return err
	}

	// Bring up the bridge interface
	if err := RunCommand("ip", "link", "set", name, "up"); err != nil {
		return err
	}

	return nil
}
