package init_app

import (
	"fmt"
	"log"
	"os/exec"
)

func CreateDisk(diskDir string, vmNumber int, diskSize int) error {
	diskPath := fmt.Sprintf("%s/disk%d.ext4", diskDir, vmNumber)
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(`dd if=/dev/zero of="%s" conv=sparse bs=1M count=%d`,
			diskPath, diskSize),
	)

	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create disk: %w", err)
	}

	return nil
}
