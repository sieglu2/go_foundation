package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sieglu2/go_foundation/foundation"
)

const (
	claudeSecretAccountName string = "my_anthropic"
	claudeSecretServiceName string = "claude"
)

type ClaudeClient struct {
	apiKey    string
	maxTokens int
	client    *http.Client
	model     string
}

type claudeContent struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Source    string `json:"source,omitempty"`
	MediaType string `json:"media_type,omitempty"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content []claudeContent `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	return NewClaudeClientWithConfig(apiKey, DefaultMaxTokens, "claude-3-opus-20240229")
}

func NewClaudeClientWithConfig(apiKey string, maxTokens int, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:    apiKey,
		maxTokens: maxTokens,
		model:     model,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func convertToClaudeMessages(messages []LlmMessage) []claudeMessage {
	claudeMessages := make([]claudeMessage, 0, len(messages))

	for _, msg := range messages {
		claudeMsg := claudeMessage{
			Role:    string(msg.Role),
			Content: make([]claudeContent, 0),
		}

		// Add text content if present
		if msg.Content != "" {
			claudeMsg.Content = append(claudeMsg.Content, claudeContent{
				Type: "text",
				Text: msg.Content,
			})
		}

		// Add image content if present
		if msg.B64Image != "" {
			claudeMsg.Content = append(claudeMsg.Content, claudeContent{
				Type:      "image",
				Source:    fmt.Sprintf("data:image/png;base64,%s", msg.B64Image),
				MediaType: "image/png",
			})
		}

		claudeMessages = append(claudeMessages, claudeMsg)
	}

	return claudeMessages
}

func (c *ClaudeClient) ReplyMessage(
	ctx context.Context,
	messages []LlmMessage,
) (string, error) {
	logger := foundation.Logger()

	if len(messages) == 0 {
		logger.Errorf("empty messages array")
		return "", fmt.Errorf("empty messages array")
	}

	systemPrompt := ""
	if messages[0].Role == RoleSystem {
		systemPrompt = messages[0].Content
		messages = messages[1:]
	}

	claudeMessages := convertToClaudeMessages(messages)

	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages:  claudeMessages,
		System:    systemPrompt,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.Errorf("failed to Marshal: %v", err)
		return "", fmt.Errorf("failed to Marshal: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		logger.Errorf("failed to NewRequestWithContext: %v", err)
		return "", fmt.Errorf("failed to NewRequestWithContext: %v", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Errorf("failed to client.Do: %v", err)
		return "", fmt.Errorf("failed to client.Do: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("failed to io.ReadAll: %v", err)
		return "", fmt.Errorf("failed to io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp claudeResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			logger.Errorf("failed to Unmarshal claudeResponse: %v", err)
			return "", fmt.Errorf("failed to parse error response, status code: %d", resp.StatusCode)
		}
		if errorResp.Error != nil {
			logger.Errorf("claude API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
			return "", fmt.Errorf("claude API error: %s - %s", errorResp.Error.Type, errorResp.Error.Message)
		}
		logger.Errorf("unexpected status code: %d", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		logger.Errorf("failed to Unmarshal claudeResponse: %v", err)
		return "", fmt.Errorf("failed to Unmarshal claudeResponse: %v", err)
	}

	if len(claudeResp.Content) == 0 {
		logger.Errorf("empty response from Claude API")
		return "", fmt.Errorf("empty response from Claude API")
	}

	return claudeResp.Content[0].Text, nil
}
