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
	deepseekSecretAccountName string = "my_deepseek"
	deepseekSecretServiceName string = "deepseek"
)

type DeepseekClient struct {
	apiKey    string
	maxTokens int
	client    *http.Client
	model     string
}

type deepseekMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// Update: Changed from ImageUrl to ImageBase64 to match API expectations
	ImageBase64 string `json:"image_base64,omitempty"`
}

type deepseekMessage struct {
	Role    string                   `json:"role"`
	Content []deepseekMessageContent `json:"content"`
}

type deepseekRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	Messages  []deepseekMessage `json:"messages"`
}

type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func NewDeepseekClient(apiKey string) *DeepseekClient {
	return NewDeepseekClientWithConfig(apiKey, DefaultMaxTokens, "deepseek-chat")
}

func NewDeepseekClientWithConfig(apiKey string, maxTokens int, model string) *DeepseekClient {
	return &DeepseekClient{
		apiKey:    apiKey,
		maxTokens: maxTokens,
		model:     model,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func convertToDeepseekMessages(messages []LlmMessage) []deepseekMessage {
	logger := foundation.Logger()
	deepseekMessages := make([]deepseekMessage, 0, len(messages))

	for _, msg := range messages {
		deepMsg := deepseekMessage{
			Role:    string(msg.Role),
			Content: make([]deepseekMessageContent, 0),
		}

		// Add text content if present
		if msg.Content != "" {
			deepMsg.Content = append(deepMsg.Content, deepseekMessageContent{
				Type: "text",
				Text: msg.Content,
			})
		}

		// Add image content if present
		if msg.B64Image != "" {
			logger.Warnf("deepseek does not accept image as input for now. 01/18/2025")
			// deepMsg.Content = append(deepMsg.Content, deepseekMessageContent{
			// 	Type:        "image",
			// 	ImageBase64: msg.B64Image, // Update: Directly use base64 string without data URI prefix
			// })
		}

		deepseekMessages = append(deepseekMessages, deepMsg)
	}

	return deepseekMessages
}

func (d *DeepseekClient) ReplyMessage(
	ctx context.Context,
	messages []LlmMessage,
) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("empty messages array")
	}

	deepseekMessages := convertToDeepseekMessages(messages)

	reqBody := deepseekRequest{
		Model:     d.model,
		MaxTokens: d.maxTokens,
		Messages:  deepseekMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to json.Marshal: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to NewRequestWithContext: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to client.Do: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp deepseekResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", fmt.Errorf("failed to parse error response, status code: %d", resp.StatusCode)
		}
		if errorResp.Error != nil {
			return "", fmt.Errorf("deepseek API error: %s - %s (code: %s)",
				errorResp.Error.Type, errorResp.Error.Message, errorResp.Error.Code)
		}
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var deepseekResp deepseekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return "", fmt.Errorf("failed to json.Unmarshal: %v", err)
	}

	if len(deepseekResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from Deepseek API")
	}

	return deepseekResp.Choices[0].Message.Content, nil
}
