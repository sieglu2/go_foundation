package llm

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/sieglu2/go_foundation/foundation"
	"google.golang.org/api/option"
)

const (
	geminiSecretAccountName string = "my_google"
	geminiSecretServiceName string = "gemini"

	GEMINI_EMBEDDINGS_MAX_TOKEN = 2048

	defaultGeminiModel = "gemini-1.5-pro"
)

type GeminiClient struct {
	client    *genai.Client
	maxTokens int32
	model     string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	return NewGeminiClientWithConfig(ctx, apiKey, int32(DefaultMaxTokens), defaultGeminiModel)
}

func NewGeminiClientWithConfig(ctx context.Context, apiKey string, maxTokens int32, model string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiClient{
		client:    client,
		maxTokens: maxTokens,
		model:     model,
	}, nil
}

func (m *GeminiClient) Close() error {
	return m.client.Close()
}

func (g *GeminiClient) ReplyMessage(
	ctx context.Context, llmMessages []LlmMessage,
) (string, error) {
	logger := foundation.Logger()

	if len(llmMessages) == 0 {
		logger.Errorf("empty messages array")
		return "", fmt.Errorf("empty messages array")
	}
	currentMessage := llmMessages[len(llmMessages)-1]
	if currentMessage.Role != RoleUser {
		logger.Errorf("last message must have the user role")
		return "", fmt.Errorf("last message must have the user role")
	}

	// manually get the last message out as the current message to fit into gemini's API mechanism.
	llmMessages = llmMessages[:len(llmMessages)-1]
	currentContent, err := convertToContent(currentMessage)
	if err != nil {
		logger.Errorf("failed to convert LlmMessage to genai.Content: %v", err)
		return "", fmt.Errorf("failed to convert LlmMessage to genai.Content: %v", err)
	}

	contents, err := convertToGeminiContents(llmMessages)
	if err != nil {
		logger.Errorf("failed to convert to Gemini contents: %v", err)
		return "", fmt.Errorf("failed to convert to Gemini contents: %v", err)
	}

	model := g.client.GenerativeModel(g.model)
	model.SetMaxOutputTokens(g.maxTokens)
	chat := model.StartChat()

	// manually overwrite the history with manual saved history
	chat.History = contents

	logger.Infof("sending request to Gemini model: %s", g.model)
	resp, err := chat.SendMessage(ctx, currentContent.Parts...)
	if err != nil {
		logger.Errorf("failed to generate content: %v", err)
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		logger.Errorf("empty response from Gemini")
		return "", errors.New("empty response from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		logger.Errorf("empty content in response")
		return "", errors.New("empty content in response")
	}

	textResponse := ""
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			textResponse += string(textPart)
		}
	}

	logger.Infof("received Gemini response")
	return textResponse, nil
}

func convertToGeminiContents(llmMessages []LlmMessage) ([]*genai.Content, error) {
	logger := foundation.Logger()

	if len(llmMessages) == 0 {
		logger.Errorf("empty llmMessages")
		return nil, fmt.Errorf("empty llmMessages")
	}

	contents := make([]*genai.Content, 0, len(llmMessages))

	for i, message := range llmMessages {
		if len(message.Content) == 0 && len(message.B64Image) == 0 {
			logger.Errorf("message %d has no content", i)
			return nil, fmt.Errorf("message %d has no content", i)
		}

		content, err := convertToContent(message)
		if err != nil {
			logger.Errorf("failed to convert LlmMessage to genai.Content: %v", err)
			return nil, fmt.Errorf("failed to convert LlmMessage to genai.Content: %v", err)
		}

		contents = append(contents, content)
	}

	return contents, nil
}

func convertToContent(message LlmMessage) (*genai.Content, error) {
	logger := foundation.Logger()

	content := &genai.Content{}

	switch message.Role {
	case RoleUser:
		content.Role = "user"
	case RoleAssistant:
		content.Role = "model"
	case RoleSystem:
		// Gemini handles system messages differently - we'll add it as a user message
		// with a special prefix, or you could add it to the first user message
		content.Role = "user"
	default:
		logger.Errorf("unknown role: %s", message.Role)
		return nil, fmt.Errorf("unknown role: %s", message.Role)
	}

	if len(message.Content) > 0 {
		content.Parts = append(content.Parts, genai.Text(message.Content))
	}

	if len(message.B64Image) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(message.B64Image)
		if err != nil {
			logger.Errorf("failed to base64 decode png bytes: %v", err)
			return nil, err
		}
		imagePart := genai.ImageData("png", []byte(decoded))
		content.Parts = append(content.Parts, imagePart)
	}
	return content, nil
}
