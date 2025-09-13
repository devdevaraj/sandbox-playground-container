package configure_network

import (
	"log"
	"strconv"

	"github.com/devdevaraj/firestarter/create_bridge"
	"github.com/devdevaraj/firestarter/setup_nat"
	"github.com/devdevaraj/firestarter/tap_interface"
)

func ConfigureNetWork(ipAdd string, ipAddB string, vms int) {
	// Create bridge
	if err := create_bridge.CreateBridge(ipAdd); err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Create TAP interfaces (e.g., tap0 and tap1)
	// tapInterfaces := []string{"tap0", "tap1"}
	for i := range vms {
		if err := tap_interface.CreateTapInterface("tap" + strconv.Itoa(i)); err != nil {
			log.Fatalf("Failed to create TAP interface %s: %v", "tap"+strconv.Itoa(i), err)
		}
	}

	// Set up NAT
	if err := setup_nat.SetupNAT(ipAddB); err != nil {
		log.Fatalf("Failed to set up NAT: %v", err)
	}

	log.Println("Network setup completed successfully.")
}
