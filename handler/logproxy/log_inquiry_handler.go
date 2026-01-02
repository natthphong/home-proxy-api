package logproxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/cache"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.uber.org/zap"
	"strings"
	"time"
)

type InquiryLogRequest struct {
	RequestId string `json:"requestId"`
	Type      string `json:"type"`
}
type InquiryLogResponse struct {
	RequestId     string          `json:"requestId"`
	EffectiveDate time.Time       `json:"effectiveDate"`
	Type          string          `json:"type"`
	Body          json.RawMessage `json:"body"`
}

type WebHookPayload struct {
	Date time.Time `json:"date"`
}

func NewLogInquiryHandler(
	getRedisFunc cache.GetRedisFunc,
	db *pgxpool.Pool,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger := logz.NewLogger()

		var (
			req InquiryLogRequest
		)
		res := []InquiryLogResponse{}
		err := c.BodyParser(&req)
		if err != nil {
			return api.BadRequest(c, "invalid request body")
		}

		if req.RequestId == "" {
			return api.BadRequest(c, "requestId is required")
		}
		if req.Type == "webhook" {
			bodyStr, _ := getRedisFunc(c.Context(), fmt.Sprintf(cache.KeyWebhook, req.RequestId))
			bodyJson := json.RawMessage(bodyStr)
			payload := WebHookPayload{}
			if err = json.Unmarshal(bodyJson, &payload); err != nil {
				logger.Error(err.Error(), zap.String("requestId", req.RequestId))
			}
			if bodyStr != "" {
				temp := InquiryLogResponse{
					EffectiveDate: payload.Date,
					RequestId:     req.RequestId,
					Type:          "webhook",
					Body:          bodyJson,
				}
				res = append(res, temp)
				return api.Ok(c, res)
			}
		}

		sql := `
			select effective_date,request_id,proxy_type,body from tbl_proxy_log
			where request_id = $1
			and ( $2 ='' or  $2  = LOWER(proxy_type) )
			`

		rows, err := db.Query(c.Context(), sql, req.RequestId, strings.ToLower(req.Type))
		if err != nil {
			logger.Error(err.Error(), zap.String("requestId", req.RequestId))
			return api.InternalError(c, "Something went wrong")
		}
		for rows.Next() {
			temp := InquiryLogResponse{}
			tempBody := ""
			err = rows.Scan(
				&temp.EffectiveDate,
				&temp.RequestId,
				&temp.Type,
				&tempBody,
			)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", req.RequestId))
				return api.InternalError(c, "Something went wrong")
			}
			tempBodyByte, err := base64.StdEncoding.DecodeString(tempBody)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", req.RequestId))
				return api.InternalError(c, "Something went wrong")
			}
			err = json.Unmarshal(tempBodyByte, &temp.Body)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", req.RequestId))
				return api.InternalError(c, "Something went wrong")
			}
			res = append(res, temp)
		}
		return api.Ok(c, res)
	}
}
