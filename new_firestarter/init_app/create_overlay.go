package init_app

import (
	"fmt"
	"log"
	"os/exec"
)

func CreateOverlay(fsDir string, vmNumber int, overlaySize int) error {
	overlayPath := fmt.Sprintf("%s/vm%d-overlay.ext4", fsDir, vmNumber)
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(`dd if=/dev/zero of="%s" conv=sparse bs=1M count=%d && mkfs.ext4 "%s"`,
			overlayPath, overlaySize, overlayPath),
	)

	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create overlay: %w", err)
	}

	return nil
}
