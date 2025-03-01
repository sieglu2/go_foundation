//go:build darwin

package secret_key

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sieglu2/go_foundation/foundation"
)

func GetSecretKey(ctx context.Context, accountName, serviceName string) (string, error) {
	logger := foundation.Logger()

	cmd := exec.CommandContext(ctx, "security", "find-generic-password",
		"-a", accountName, "-s", serviceName, "-w")

	output, err := cmd.Output()
	if err != nil {
		logger.Infof("failed to CommandContext for %s/%s: %v", accountName, serviceName, err)
		return "", fmt.Errorf("failed to get secret for %s/%s: %v",
			accountName, serviceName, err)
	}

	return strings.TrimSpace(string(output)), nil
}
