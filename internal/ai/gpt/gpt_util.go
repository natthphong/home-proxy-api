package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai"
	"go.uber.org/zap"
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

type SendTextAndGetTextFunc func(ctx context.Context, logger *zap.Logger, req ai.PromptRequest) (string, error)

func SendTextAndGetText(client *openai.Client) SendTextAndGetTextFunc {
	return func(ctx context.Context, logger *zap.Logger, req ai.PromptRequest) (string, error) {
		msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.ChatHistory)+1)
		for i, h := range req.ChatHistory {
			role := strings.ToLower(strings.TrimSpace(h.Role))
			content := strings.TrimSpace(h.Message)
			if content == "" {
				continue
			}
			switch role {
			case ai.RoleSystem:
				msgs = append(msgs, openai.SystemMessage(content))
			case ai.RoleAssistant:
				msgs = append(msgs, openai.AssistantMessage(content))
			case ai.RoleUser, "":
				msgs = append(msgs, openai.UserMessage(content))
			default:
				return "", fmt.Errorf("chatHistory[%d]: invalid role %q (allowed: system/user/assistant)", i, h.Role)
			}
		}

		if strings.TrimSpace(req.Prompt) != "" {
			msgs = append(msgs, openai.UserMessage(req.Prompt))
		}
		if len(msgs) == 0 {
			return "", fmt.Errorf("no messages to send (prompt empty and chatHistory empty)")
		}

		params := openai.ChatCompletionNewParams{
			Messages: openai.F(msgs),
			Model:    openai.F(openai.ChatModelGPT4oMini),
		}

		if req.ResponseOptions != nil {
			ro := req.ResponseOptions

			name := strings.TrimSpace(ro.Name)
			if name == "" {
				name = "response"
			}

			// Schema must be a JSON object (your RawMessage should contain: {"type":"object",...})
			if len(ro.Schema) == 0 {
				return "", fmt.Errorf("responseOptions.schema is required when responseOptions is set")
			}

			var schemaAny any
			if err := json.Unmarshal(ro.Schema, &schemaAny); err != nil {
				return "", fmt.Errorf("invalid responseOptions.schema json: %w", err)
			}

			rf := openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F(name),
					Description: openai.F(ro.Description),
					Schema:      openai.F(schemaAny),
					Strict:      openai.F(ro.Strict),
				}),
			}
			params.ResponseFormat = openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](rf)

		}

		chatCompletion, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			logger.Error("chat completion failed", zap.Error(err))
			return "", err
		}
		if len(chatCompletion.Choices) == 0 {
			return "", fmt.Errorf("no choices returned")
		}

		return chatCompletion.Choices[0].Message.Content, nil
	}
}

type SendTextAndGetAudioFunc func(ctx context.Context, logger *zap.Logger, prompt string) ([]byte, error)

func SendTextAndGetAudio(client *openai.Client) SendTextAndGetAudioFunc {
	return func(ctx context.Context, logger *zap.Logger, prompt string) ([]byte, error) {
		resp, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
			Input:          openai.F(prompt),
			Model:          openai.F("gpt-4o-mini-tts-2025-03-20"),
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

type SendAudioAndGetTextFunc func(
	ctx context.Context,
	logger *zap.Logger,
	audio []byte,
	filename string,
	contentType string,
) (string, error)

func SendAudioAndGetText(client *openai.Client) SendAudioAndGetTextFunc {
	return func(
		ctx context.Context,
		logger *zap.Logger,
		audio []byte,
		filename string,
		contentType string,
	) (string, error) {

		if len(audio) == 0 {
			return "", fmt.Errorf("audio is empty")
		}
		if filename == "" {
			filename = "audio.wav"
		}
		if contentType == "" {
			contentType = "audio/wav"
		}

		resp, err := client.Audio.Transcriptions.New(
			ctx,
			openai.AudioTranscriptionNewParams{
				File: openai.FileParam(
					bytes.NewReader(audio),
					filename,
					contentType,
				),

				Model: openai.F("gpt-4o-mini-transcribe"),
			},
		)
		if err != nil {
			logger.Error("stt failed", zap.Error(err))
			return "", err
		}

		return resp.Text, nil
	}
}

type SendTextAndGetAudioWithStyleFunc func(
	ctx context.Context,
	logger *zap.Logger,
	text string,
	style string,
	speed float32,
) ([]byte, error)

func SendTextAndGetAudioWithStyle(client *openai.Client) SendTextAndGetAudioWithStyleFunc {
	return func(
		ctx context.Context,
		logger *zap.Logger,
		text string,
		style string,
		speed float32,
	) ([]byte, error) {

		s := float64(speed)
		if s <= 0 {
			s = 1.0
		}
		if s < 0.25 {
			s = 0.25
		}
		if s > 4.0 {
			s = 4.0
		}

		prompt := text
		if style != "" {
			prompt = "Style: " + style + "\n\nText: " + text
		}

		resp, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
			Input:          openai.F(prompt),
			Model:          openai.F("gpt-4o-mini-tts"),
			Voice:          openai.F(openai.AudioSpeechNewParamsVoiceAlloy),
			ResponseFormat: openai.F(openai.AudioSpeechNewParamsResponseFormatMP3),
			Speed:          openai.F(s),
		})
		if err != nil {
			logger.Error("tts failed", zap.Error(err))
			return nil, err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("tts read failed", zap.Error(err))
			return nil, err
		}
		return b, nil
	}
}
