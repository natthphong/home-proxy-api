package ai_proxy

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai"
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
			req ai.PromptRequest
			res string
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
		if model != Gpt && (len(req.ChatHistory) > 0 || req.ResponseOptions != nil) {
			return api.BadRequest(c, "support function only gpt")
		}
		if model == Gpt {
			modelUse = Gpt
			logger.Debug("gptPrompt", zap.String("reqID", reqID))
			res, err = gptPrompt(ctx, logger, req)
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
		resp := fiber.Map{
			"result": res,
			"model":  modelUse,
		}
		if req.ResponseOptions != nil {
			var resJson interface{}
			err = json.Unmarshal([]byte(res), &resJson)
			if err != nil {
				return api.InternalError(c, "Something went wrong")
			}
			resp["resultJson"] = resJson
		}
		return api.Ok(c, resp)
	}
}
