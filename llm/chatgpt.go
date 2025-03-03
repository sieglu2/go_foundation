package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sieglu2/go_foundation/foundation"
)

const (
	chatgptSecretAccountName string = "my_openai"
	chatgptSecretServiceName string = "chatgpt"
)

type ChatGptClient struct {
	client    *openai.Client
	maxTokens int
	model     string
}

func NewChatGptClient(apiKey string) *ChatGptClient {
	return NewChatGptClientWithConfig(apiKey, DefaultMaxTokens, "gpt-4-turbo")
}

func NewChatGptClientWithConfig(apiKey string, maxTokens int, model string) *ChatGptClient {
	openaiClient := openai.NewClient(apiKey)
	return &ChatGptClient{
		client:    openaiClient,
		maxTokens: maxTokens,
		model:     model,
	}
}

func (t *ChatGptClient) Close() error {
	return nil
}

func convertToChatGptMessages(chatGptMessages []LlmMessage) ([]openai.ChatCompletionMessage, error) {
	logger := foundation.Logger()

	if len(chatGptMessages) == 0 {
		logger.Errorf("empty chatGptMessages")
		return nil, fmt.Errorf("empty chatGptMessages")
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(chatGptMessages))
	for i, chatGptMessage := range chatGptMessages {
		if len(chatGptMessage.Content) == 0 && len(chatGptMessage.B64Image) == 0 {
			logger.Errorf("message %d has no content", i)
			return nil, fmt.Errorf("message %d has no content", i)
		}

		chatCompletionMessage := openai.ChatCompletionMessage{
			Role: string(chatGptMessage.Role),
		}

		if len(chatGptMessage.B64Image) > 0 {
			// Ensure there's always a valid text content, even if empty
			textContent := chatGptMessage.Content
			if textContent == "" {
				textContent = " " // Provide a non-empty string if content is empty
			}

			chatCompletionMessage.MultiContent = []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: textContent,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: fmt.Sprintf("data:image/png;base64,%s", chatGptMessage.B64Image),
					},
				},
			}
		} else {
			chatCompletionMessage.Content = chatGptMessage.Content
		}
		messages = append(messages, chatCompletionMessage)
	}

	return messages, nil
}

func (t *ChatGptClient) ReplyMessage(
	ctx context.Context, llmMessages []LlmMessage,
) (string, error) {
	logger := foundation.Logger()

	messages, err := convertToChatGptMessages(llmMessages)
	if err != nil {
		logger.Errorf("failed to convertToChatGptMessages: %v", err)
		return "", fmt.Errorf("failed to convertToChatGptMessages: %v", err)
	}

	request := openai.ChatCompletionRequest{
		Model:     t.model,
		MaxTokens: t.maxTokens,
		Messages:  messages,
	}

	logger.Infof("sending request %+v to ChatGpt", request)
	resp, err := t.client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Errorf("failed to CreateChatCompletion: %v", err)
		return "", fmt.Errorf("failed to CreateChatCompletion: %v", err)
	}

	if len(resp.Choices) == 0 {
		logger.Errorf("empty resp.Choices")
		return "", fmt.Errorf("empty resp.Choices")
	}

	logger.Infof("receive ChatGpt response.")
	return resp.Choices[0].Message.Content, nil
}
