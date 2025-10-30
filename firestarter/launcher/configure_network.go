package launcher

import (
	"log"

	"github.com/devdevaraj/firestarter/init_app"
)

func ConfigureNetWork(cfg init_app.Config) {
	// Create bridge
	if err := CreateBridge(*cfg.BridgeIP, *cfg.Bridge); err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Create TAP interfaces (e.g., tap0 and tap1)
	for i := range len(cfg.Templates) {
		for _, j := range cfg.Templates[i].Network {
			if err := CreateTapInterface(j.TAP, j.Bridge); err != nil {
				log.Fatalf("Failed to create TAP interface %s: %v", j.TAP, err)
			}
		}
	}

	// Set up NAT
	if err := SetupNAT(*cfg.Network, *cfg.Bridge); err != nil {
		log.Fatalf("Failed to set up NAT: %v", err)
	}

	log.Println("Network setup completed successfully.")
}
