package ai_proxy

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/cache"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.uber.org/zap"
)

func NewChatModelAllHandler(
	db *pgxpool.Pool,
	getRedisFunc cache.GetRedisFunc,
	setRedisFunc cache.SetRedisFunc,
) fiber.Handler {

	return func(c *fiber.Ctx) error {

		ctx := c.Context()
		logger := logz.NewLogger()
		modelList := []string{}
		reqID := c.Get("requestId")
		redisStrResponse, err := getRedisFunc(ctx, cache.KeyAiModel)
		if err != nil || redisStrResponse == "" {
			logger.Debug("no data in redis cache")
			sql := `
			select model  from tbl_ai_model
			where is_deleted = 'N'
			`
			rows, err := db.Query(ctx, sql)
			for rows.Next() {
				var temp string
				err = rows.Scan(&temp)
				if err != nil {
					logger.Error(err.Error(), zap.String("requestId", reqID))
					return api.InternalError(c, "Something went wrong")
				}
				modelList = append(modelList, temp)
			}

			redisStrModelList, err := json.Marshal(modelList)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", reqID))
				return api.InternalError(c, "Something went wrong")
			}
			err = setRedisFunc(ctx, cache.KeyAiModel, redisStrModelList, 0)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", reqID))
				return api.InternalError(c, "Something went wrong")
			}

		} else {
			err = json.Unmarshal([]byte(redisStrResponse), &modelList)
			if err != nil {
				logger.Error(err.Error(), zap.String("requestId", reqID))
				return api.InternalError(c, "Something went wrong")
			}
		}

		return api.Ok(c, fiber.Map{
			"modelList": modelList,
		})
	}
}
