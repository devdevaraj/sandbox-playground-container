package launcher

import (
	"fmt"
	"os/exec"
)

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running %s %v: %v\nOutput: %s", name, args, err, output)
	}
	return nil
}
