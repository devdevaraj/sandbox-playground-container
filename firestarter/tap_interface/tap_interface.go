package tap_interface

import "github.com/devdevaraj/firestarter/run_command"

func CreateTapInterface(tapName string) error {
	// Delete existing TAP interface if it exists
	run_command.RunCommand("ip", "link", "set", tapName, "down")
	run_command.RunCommand("ip", "link", "delete", tapName)

	// Create TAP interface
	if err := run_command.RunCommand("ip", "tuntap", "add", "dev", tapName, "mode", "tap"); err != nil {
		return err
	}

	// Attach TAP interface to bridge
	if err := run_command.RunCommand("ip", "link", "set", tapName, "master", "br0"); err != nil {
		return err
	}

	// Bring up the TAP interface
	if err := run_command.RunCommand("ip", "link", "set", tapName, "up"); err != nil {
		return err
	}

	return nil
}
