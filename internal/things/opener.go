package things

import (
	"fmt"
	"os/exec"
)

// OpenURL opens a URL using macOS open command
func OpenURL(url string) error {
	cmd := exec.Command("open", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to open URL: %s: %w", string(output), err)
	}
	return nil
}
