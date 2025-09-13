package create_bridge

import (
	run_command "github.com/devdevaraj/firestarter/run_command"
)

func CreateBridge(ipAddr string) error {
	// Delete existing bridge if it exists
	run_command.RunCommand("ip", "link", "set", "br0", "down")
	run_command.RunCommand("ip", "link", "delete", "br0", "type", "bridge")

	// Create bridge
	if err := run_command.RunCommand("ip", "link", "add", "name", "br0", "type", "bridge"); err != nil {
		return err
	}

	// Assign IP address to bridge
	if err := run_command.RunCommand("ip", "addr", "add", ipAddr, "dev", "br0"); err != nil {
		return err
	}

	// Bring up the bridge interface
	if err := run_command.RunCommand("ip", "link", "set", "br0", "up"); err != nil {
		return err
	}

	return nil
}
