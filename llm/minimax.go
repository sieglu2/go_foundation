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
	minimaxApiEndpoint = "https://api.minimaxi.chat/v1/text/chatcompletion_v2"

	minimaxSecretAccountName string = "my_hailuoai"
	minimaxSecretServiceName string = "minimax"
)

type MinimaxClient struct {
	apiKey    string
	maxTokens int
	model     string
	client    *http.Client
}

// Message structs to handle both formats
type MinimaxMessage struct {
	Role    string      `json:"role"`
	Name    string      `json:"name,omitempty"`
	Content interface{} `json:"content"` // Can be string or []MinimaxContent
}

type MinimaxContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *MinimaxImageURL `json:"image_url,omitempty"`
}

type MinimaxImageURL struct {
	URL string `json:"url"`
}

type MinimaxRequest struct {
	Model    string           `json:"model"`
	Messages []MinimaxMessage `json:"messages"`
}

type MinimaxChoice struct {
	Message struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
	Index        int    `json:"index"`
}

type MinimaxResponse struct {
	ID      string          `json:"id"`
	Choices []MinimaxChoice `json:"choices"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Object  string          `json:"object"`
	Usage   MinimaxUsage    `json:"usage"`
}

type MinimaxUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewMinimaxClient(apiKey string) *MinimaxClient {
	return NewMinimaxClientWithConfig(apiKey, DefaultMaxTokens, "MiniMax-Text-01")
}

func NewMinimaxClientWithConfig(apiKey string, maxTokens int, model string) *MinimaxClient {
	return &MinimaxClient{
		apiKey:    apiKey,
		maxTokens: maxTokens,
		model:     model,
		client:    &http.Client{},
	}
}

func convertToMinimaxMessages(llmMessages []LlmMessage) ([]MinimaxMessage, error) {
	logger := foundation.Logger()

	if len(llmMessages) == 0 {
		logger.Errorf("empty llmMessages")
		return nil, fmt.Errorf("empty messages array")
	}

	messages := make([]MinimaxMessage, 0, len(llmMessages))
	for i, msg := range llmMessages {
		if len(msg.Content) == 0 && len(msg.B64Image) == 0 {
			logger.Errorf("message %d has no content", i)
			return nil, fmt.Errorf("message %d has no content", i)
		}

		minimaxMessage := MinimaxMessage{
			Role: string(msg.Role),
			Name: getMinimaxRoleName(msg.Role),
		}

		// Handle text-only case
		if len(msg.B64Image) == 0 {
			minimaxMessage.Content = msg.Content
		} else {
			// Handle multimedia case
			contents := []MinimaxContent{
				{
					Type: "text",
					Text: msg.Content,
				},
				{
					Type: "image_url",
					ImageURL: &MinimaxImageURL{
						URL: fmt.Sprintf("data:image/png;base64,%s", msg.B64Image),
					},
				},
			}
			minimaxMessage.Content = contents // Use the array directly as content
		}

		messages = append(messages, minimaxMessage)
	}

	return messages, nil
}

func (m *MinimaxClient) ReplyMessage(ctx context.Context, llmMessages []LlmMessage) (string, error) {
	logger := foundation.Logger()

	messages, err := convertToMinimaxMessages(llmMessages)
	if err != nil {
		logger.Errorf("failed to convertToMinimaxMessages: %v", err)
		return "", fmt.Errorf("failed to convertToMinimaxMessages: %v", err)
	}

	request := MinimaxRequest{
		Model:    m.model,
		Messages: messages,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		logger.Errorf("failed to marshal request: %v", err)
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", minimaxApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Errorf("failed to create request: %v", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	resp, err := m.client.Do(req)
	if err != nil {
		logger.Errorf("failed to send request: %v", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("received non-200 status code: %d, body: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var minimaxResponse MinimaxResponse
	if err := json.Unmarshal(body, &minimaxResponse); err != nil {
		logger.Errorf("failed to unmarshal response: %v", err)
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	logger.Infof("received MiniMax response")

	if len(minimaxResponse.Choices) == 0 {
		logger.Errorf("empty resp.Choices")
		return "", fmt.Errorf("no completion choices returned")
	}

	return minimaxResponse.Choices[0].Message.Content, nil
}

func getMinimaxRoleName(role LlmRole) string {
	switch role {
	case RoleSystem:
		return "MM Intelligent Assistant"
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	}
	return ""
}
