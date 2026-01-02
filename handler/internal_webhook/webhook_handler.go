package internal_webhook

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/cache"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/kafka"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.uber.org/zap"
	"time"
)

type WebHookRequest struct {
	Date    *time.Time      `json:"date,omitempty"`
	Body    json.RawMessage `json:"body"`
	Channel string          `json:"channel"`
}

type KafkaWebHookProduce struct {
	RequestId string          `json:"requestId"`
	Date      time.Time       `json:"date"`
	Body      json.RawMessage `json:"body"`
	Channel   string          `json:"channel"`
}

func NewWebHookHandler(
	sendMessageKafka kafka.SendMessageSyncFunc,
	setFunc cache.SetRedisFunc,
	db *pgxpool.Pool,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()
		var req WebHookRequest
		ctx := c.Context()
		err := c.BodyParser(&req)
		if err != nil {
			return api.BadRequest(c, "invalid request body")
		}
		if req.Channel == "" {
			return api.BadRequest(c, "channel is required")
		}
		date := time.Now()
		if req.Date != nil {
			date = *req.Date
		}
		reqId := c.Get("requestId")
		messageProduce := &KafkaWebHookProduce{
			RequestId: reqId,
			Date:      date,
			Body:      req.Body,
			Channel:   req.Channel,
		}
		bodyBytes, err := json.Marshal(messageProduce)
		if err != nil {
			return err
		}
		bodyEnc := base64.StdEncoding.EncodeToString(bodyBytes)
		_, err = db.Exec(c.Context(),
			`INSERT INTO tbl_proxy_log (effective_date, request_id, proxy_type, body)
			 VALUES ($1, $2, $3, $4)`,
			messageProduce.Date,
			reqId,
			"webhook",
			bodyEnc,
		)
		if err != nil {
			return api.InternalError(c, "failed to insert db")
		}
		err = sendMessageKafka(logger, messageProduce)
		if err != nil {
			return api.InternalError(c, "failed to send message")
		}
		err = setFunc(ctx, fmt.Sprintf(cache.KeyWebhook, reqId), string(bodyBytes), time.Hour*24)
		if err != nil {
			logger.Error(err.Error(), zap.String("requestId", reqId))
		}
		return api.Ok(c, nil)
	}

}
