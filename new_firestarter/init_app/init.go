package init_app

import (
	"fmt"
	"log"
	"os"
)

func Init() (*Config, []string) {
	args := os.Args
	cfg := ParseConfigs(args[1])

	fsDir := "/root/firecracker/overlayfs"
	err := os.MkdirAll(fsDir, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return nil, nil
	}

	for i, v := range cfg.Templates {
		if *v.EnableOverlay {
			if err := CreateOverlay(fsDir, i+1, *v.OverlaySize); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	}

	return &cfg, args
}
