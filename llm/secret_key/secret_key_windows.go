//go:build windows

package secret_key

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func GetSecretKey(ctx context.Context, accountName, serviceName string) (string, error) {
	// Define known account/service pairs
	knownPairs := map[string]struct {
		account string
		service string
	}{
		"chatgpt": {
			account: "my_openai",
			service: "chatgpt",
		},
		"claude": {
			account: "my_anthropic",
			service: "claude",
		},
		"deepseek": {
			account: "my_deepseek",
			service: "deepseek",
		},
	}

	// Find which provider this account/service pair belongs to
	var provider string
	for p, pair := range knownPairs {
		if pair.account == accountName && pair.service == serviceName {
			provider = p
			break
		}
	}

	if provider == "" {
		return "", fmt.Errorf("unknown account/service pair: %s/%s", accountName, serviceName)
	}

	// Get the environment variable
	envVarName := strings.ToUpper(provider) + "_API_KEY"
	key := os.Getenv(envVarName)
	if key == "" {
		return "", fmt.Errorf("no API key found in environment variable %s", envVarName)
	}

	return key, nil
}
