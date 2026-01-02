package ai_proxy

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/gemini"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/gpt"
	meta_ai "gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/metaai"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.uber.org/zap"
)

const (
	Gpt    = "gpt"
	Gemini = "gemini"
	Meta   = "meta"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

func NewPromptHandler(
	geminiPrompt gemini.SendTextAndGetTextFunc,
	gptPrompt gpt.SendTextAndGetTextFunc,
	metaClient *meta_ai.MetaAI,

) fiber.Handler {

	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		ctx := c.Context()
		reqID := c.Get("requestId")
		var (
			req PromptRequest
			res interface{}
		)
		err := c.BodyParser(&req)
		if err != nil {
			return api.BadRequest(c, "invalidate prompt request")
		}
		if req.Prompt == "" {
			return api.BadRequest(c, "prompt is required")
		}
		modelUse := Gemini
		model := strings.ToLower(req.Model)
		if model == Gpt {
			modelUse = Gpt
			logger.Debug("gptPrompt", zap.String("reqID", reqID))
			res, err = gptPrompt(ctx, logger, req.Prompt)
		} else if model == Meta {
			modelUse = Meta
			res, err = metaClient.Prompt(req.Prompt, false, 0, true)
		} else {
			logger.Debug("geminiPrompt", zap.String("reqID", reqID))
			res, err = geminiPrompt(ctx, logger, req.Prompt)
		}
		if err != nil {
			return api.InternalError(c, "Something went wrong")
		}

		return api.Ok(c, fiber.Map{
			"result": res,
			"model":  modelUse,
		})
	}
}
