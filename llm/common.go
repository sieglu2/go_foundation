package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sieglu2/go_foundation/foundation"
	"github.com/sieglu2/go_foundation/llm/secret_key"
)

type LlmRole string

const (
	DefaultMaxTokens int = 1000
	DefaultTimeout       = 30 * time.Second

	RoleAssistant LlmRole = "assistant"
	RoleSystem    LlmRole = "system"
	RoleUser      LlmRole = "user"
)

type LlmMessage struct {
	Role     LlmRole `json:"role"`
	Content  string  `json:"content"`
	B64Image string
}

type LlmClient interface {
	ReplyMessage(ctx context.Context, messages []LlmMessage) (string, error)
	Close() error
}

func NewLlmClient() (LlmClient, error) {
	logger := foundation.Logger()
	var errors []error

	claudeApiKey, err := getSecretKey(claudeSecretAccountName, claudeSecretServiceName)
	if err == nil {
		logger.Infof("using claude client")
		return NewClaudeClient(claudeApiKey), nil
	}
	errors = append(errors, fmt.Errorf("claude client init failed: %w", err))

	chatgptApiKey, err := getSecretKey(chatgptSecretAccountName, chatgptSecretServiceName)
	if err == nil {
		logger.Infof("using chatgpt client")
		return NewChatGptClient(chatgptApiKey), nil
	}
	errors = append(errors, fmt.Errorf("chatgpt client init failed: %w", err))

	geminiApiKey, err := getSecretKey(geminiSecretAccountName, geminiSecretServiceName)
	if err == nil {
		logger.Infof("using gemini client")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		geminiClient, err := NewGeminiClient(ctx, geminiApiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create gemini client: %v", err)
		}
		return geminiClient, nil
	}

	deepseekApiKey, err := getSecretKey(deepseekSecretAccountName, deepseekSecretServiceName)
	if err == nil {
		logger.Infof("using deepseek client")
		return NewDeepseekClient(deepseekApiKey), nil
	}
	errors = append(errors, fmt.Errorf("deepseek client init failed: %w", err))

	minimaxApiKey, err := getSecretKey(minimaxSecretAccountName, minimaxSecretServiceName)
	if err == nil {
		logger.Infof("using minimax client")
		return NewMinimaxClient(minimaxApiKey), nil
	}
	errors = append(errors, fmt.Errorf("minimax client init failed: %w", err))

	return nil, fmt.Errorf("no viable client available, errors: %v", errors)
}

func getSecretKey(accountName, serviceName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key, err := secret_key.GetSecretKey(ctx, accountName, serviceName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout getting secret for %s/%s", accountName, serviceName)
		}
		return "", fmt.Errorf("failed to GetSecretKey: %v", err)
	}

	if key == "" {
		return "", fmt.Errorf("empty secret key returned")
	}

	return key, nil
}
