//go:build !darwin && !windows

package secret_key

import (
	"context"
	"fmt"
	"runtime"
)

func GetSecretKey(ctx context.Context, accountName, serviceName string) (string, error) {
	return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}
