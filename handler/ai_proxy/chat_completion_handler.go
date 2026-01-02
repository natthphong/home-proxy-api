package ai_proxy

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/revrost/go-openrouter"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.uber.org/zap"
)

type ChatCompletionRequest struct {
	ChatHistory *[]openrouter.ChatCompletionMessage `json:"chatHistory,omitempty"`
	Prompt      string                              `json:"prompt"`
	Model       *string                             `json:"model,omitempty"`
}

func NewChatCompletionHandler(
	db *pgxpool.Pool,
	client *openrouter.Client,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		reqID := c.Get("requestId")
		ctx := c.Context()
		var (
			req          ChatCompletionRequest
			defaultModel string
			messages     []openrouter.ChatCompletionMessage
		)

		err := c.BodyParser(&req)
		if err != nil {
			return api.BadRequest(c, "Invalidate Body Request")
		}
		if req.Prompt == "" {
			return api.BadRequest(c, "Prompt is required")
		}
		if req.ChatHistory != nil && len(*req.ChatHistory) > 0 {
			messages = append(messages, *req.ChatHistory...)
		}

		messages = append(messages, openrouter.ChatCompletionMessage{
			Role:    openrouter.ChatMessageRoleUser,
			Content: openrouter.Content{Text: req.Prompt},
		})
		if req.Model == nil || *req.Model == "" {
			sql := `
			select model  from tbl_ai_model
			where  is_default = 'Y'
			`
			rows, err := db.Query(ctx, sql)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", reqID))
				return api.InternalError(c, "Something went wrong")
			}
			for rows.Next() {
				err = rows.Scan(&defaultModel)
				if err != nil {
					logger.Error(err.Error(), zap.String("requestId", reqID))
					return api.InternalError(c, "Something went wrong")
				}
			}
		} else {
			defaultModel = *req.Model
			//sql := `
			//	select model  from tbl_ai_model
			//	where model = $1 and is_deleted = 'N'
			//	`
			//rows, err := db.Query(ctx, sql)
			//temp := ""
			//for rows.Next() {
			//	err = rows.Scan(&temp)
			//	if err != nil {
			//		logger.Error(err.Error(), zap.String("requestId", reqID))
			//		return api.InternalError(c, "Something went wrong")
			//	}
			//}
			//if temp == "" {
			//	return api.BadRequest(c, "Model not found")
			//}
		}

		// TODO choice model from Messages

		resp, err := client.CreateChatCompletion(
			ctx,
			openrouter.ChatCompletionRequest{
				Model:    defaultModel,
				Messages: messages,
			},
		)

		if len(resp.Choices) == 0 {
			value, _ := json.Marshal(resp)

			logger.Error(string(value), zap.String("requestId", reqID))
			return api.InternalError(c, "No Response")
		}
		return api.Ok(c, fiber.Map{
			"result": resp.Choices[0].Message.Content,
		})
	}
}
