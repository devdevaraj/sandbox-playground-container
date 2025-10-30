package launcher

import (
	"log"
	"strconv"

	"github.com/devdevaraj/firestarter/init_app"
)

func ConfigureNetWork(cfg init_app.Config) {
	// Create bridge
	if err := CreateBridge(*cfg.BridgeIP); err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Create TAP interfaces (e.g., tap0 and tap1)
	// tapInterfaces := []string{"tap0", "tap1"}
	for i := range len(cfg.Templates) {
		if err := CreateTapInterface("tap" + strconv.Itoa(i)); err != nil {
			log.Fatalf("Failed to create TAP interface %s: %v", "tap"+strconv.Itoa(i), err)
		}
	}

	// Set up NAT
	if err := SetupNAT(*cfg.Network); err != nil {
		log.Fatalf("Failed to set up NAT: %v", err)
	}

	log.Println("Network setup completed successfully.")
}
