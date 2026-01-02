package gemini

import (
	"bytes"
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/genai"
	"sync"
	"text/template"
)

var (
	availableModelsRR = []string{ // RR for Round-Robin to differentiate
		"gemini-2.0-flash",
		"gemini-1.5-flash",
		"gemini-2.0-flash-lite",
		"gemini-1.5-flash-8b",
	}
	currentModelIndex = 0
	mu                sync.Mutex // To make it safe for concurrent calls
)

func getNextModelInSequence() string {
	if len(availableModelsRR) == 0 {
		return "gemini-2.0-flash"
	}
	mu.Lock()
	model := availableModelsRR[currentModelIndex]
	currentModelIndex = (currentModelIndex + 1) % len(availableModelsRR)
	mu.Unlock()
	return model
}

func Open(ctx context.Context, apiKey string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return client, nil
}

func GeneratePrompt(templateFormat string, values map[string]string) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateFormat)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var promptBuilder bytes.Buffer
	err = tmpl.Execute(&promptBuilder, values)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return promptBuilder.String(), nil
}

type SendTextAndGetTextFunc func(ctx context.Context, logger *zap.Logger, prompt string) (string, error)

func SendTextAndGetText(client *genai.Client) SendTextAndGetTextFunc {
	return func(ctx context.Context,
		logger *zap.Logger,
		prompt string,
	) (string, error) {
		// Build the content parts
		contents := []*genai.Content{}
		parts := []*genai.Part{{Text: prompt}}
		contents = append(contents, &genai.Content{Parts: parts, Role: genai.RoleUser})
		selectedModel := getNextModelInSequence()
		logger.Debug("Selected Gemini model", zap.String("model", selectedModel))
		response, err := client.Models.GenerateContent(ctx, selectedModel, contents, nil)
		if err != nil {
			logger.Error("Gemini GenerateContent error", zap.Error(err))
			return "", err
		}

		if len(response.Text()) == 0 {
			return "", fmt.Errorf("empty response from Gemini API")
		}

		return response.Text(), nil
	}
}
