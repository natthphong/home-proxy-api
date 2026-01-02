package gpt

import (
	"bytes"
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.uber.org/zap"
	"io"
	"text/template"
)

func Open(apiKey string) (r *openai.Client) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return client
}

func GeneratePrompt(templateFormat string, values map[string]string) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateFormat)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var prompt bytes.Buffer
	err = tmpl.Execute(&prompt, values)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return prompt.String(), nil
}

type SendTextAndGetTextFunc func(ctx context.Context, logger *zap.Logger, prompt string) (string, error)

func SendTextAndGetText(client *openai.Client) SendTextAndGetTextFunc {
	return func(ctx context.Context, logger *zap.Logger, prompt string) (string, error) {
		chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			}),
			Model: openai.F(openai.ChatModelGPT4oMini),
			/*MaxTokens: openai.Int(4000),*/
		})
		if err != nil {
			return "", err
		}
		return chatCompletion.Choices[0].Message.Content, err
	}
}

type SendTextAndGetAudioFunc func(ctx context.Context, logger *zap.Logger, prompt string) ([]byte, error)

func SendTextAndGetAudio(client *openai.Client) SendTextAndGetAudioFunc {
	return func(ctx context.Context, logger *zap.Logger, prompt string) ([]byte, error) {
		resp, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
			Input:          openai.F(prompt),
			Model:          openai.F(openai.SpeechModelTTS1),
			Voice:          openai.F(openai.AudioSpeechNewParamsVoiceAlloy),
			ResponseFormat: openai.F(openai.AudioSpeechNewParamsResponseFormatMP3),
			Speed:          openai.F(0.250000),
		})
		if err != nil {
			logger.Error(err.Error())
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err.Error())
		}
		return b, nil
	}
}
