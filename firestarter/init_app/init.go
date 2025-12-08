package init_app

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func Init() *Config {
	jsonStr := os.Getenv("PG_CONFIG")
	cfg := ParseConfigs(jsonStr)

	fsDir := "/root/firecracker/overlayfs"
	err := os.MkdirAll(fsDir, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return nil
	}

	for i, v := range cfg.Templates {
		if *v.EnableOverlay {
			if err := CreateOverlay(fsDir, i+1, *v.OverlaySize); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}

		diskDir := "/root/firecracker/disks-vm" + strconv.Itoa(i+1)
		err = os.MkdirAll(diskDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return nil
		}

		for j, w := range v.Disks {
			if err := CreateDisk(diskDir, j+1, w.Size); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	}

	return &cfg
}
