package ai

import "encoding/json"

type PromptRequest struct {
	Prompt          string           `json:"prompt"`
	Model           string           `json:"model"`
	ResponseOptions *ResponseOptions `json:"responseOptions"`
	ChatHistory     []ChatHistory    `json:"chatHistory"`
}

const (
	RoleSystem    string = "system"
	RoleUser      string = "user"
	RoleAssistant string = "assistant"
)

type ChatHistory struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type ResponseOptions struct {
	Name        string          `json:"name,required"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Strict      bool            `json:"strict"`
}
