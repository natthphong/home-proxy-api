package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/revrost/go-openrouter"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/api"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/config"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/handler/ai_proxy"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/handler/email"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/handler/internal_webhook"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/handler/line"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/handler/logproxy"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/gemini"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/gpt"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/ai/metaai"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/cache"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/db"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/httputil"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/kafka"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/scramkafka"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/tracing"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/line_message"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/middleware"
	"go.uber.org/zap"
)

func main() {
	currentTime := time.Now()
	versionDeploy := currentTime.Unix()
	ctx := context.Background()
	app := initFiber()
	config.InitTimeZone()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(errors.New("unable to initial config"))
	}

	logz.Init(cfg.LogConfig.Level, cfg.Server.Name)
	defer logz.Drop()

	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	logger := zap.L()
	logger.Info("version " + strconv.FormatInt(versionDeploy, 10))
	//jsonCfg, err := json.Marshal(cfg)
	//_ = jsonCfg
	//logger.Debug("after cfg : " + string(jsonCfg))
	shutdown, err := tracing.Init(ctx, *cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdown(context.Background()) }()
	logger.Info("Otel connected")

	dbPool, err := db.Open(ctx, cfg.DBConfig)
	if err != nil {
		logger.Fatal("server connect to db", zap.Error(err))
	}
	defer dbPool.Close()
	logger.Info("DB CONNECT")

	httpClient := httputil.InitHttpClient(
		cfg.HTTP.TimeOut,
		cfg.HTTP.MaxIdleConn,
		cfg.HTTP.MaxIdleConnPerHost,
		cfg.HTTP.MaxConnPerHost,
	)
	_ = httpClient
	redisClient, err := cache.Initialize(ctx, cfg.RedisConfig)
	if err != nil {
		logger.Fatal("server connect to redis", zap.Error(err))
	}
	redisCMD := redisClient.UniversalClient()
	defer func() {
		err = redisCMD.Close()
		if err != nil {
			logger.Fatal("closing redis connection error", zap.Error(err))
		}
	}()
	logger.Info("Redis Connected")

	var botLines []*messaging_api.MessagingApiAPI
	for _, lineConf := range cfg.LineConfig {
		botLine, err := messaging_api.NewMessagingApiAPI(lineConf.ChannelToken)
		if err != nil {
			panic(err)
		}
		botLines = append(botLines, botLine)
	}
	logger.Info("line CONNECT")

	notificationFunc := line_message.SendMessage(botLines)

	clientOpenAi := gpt.Open(cfg.OpenAiConfig.ApiKey)
	clientGemini, err := gemini.Open(ctx, cfg.GemeniConfig.ApiKey)
	if err != nil {
		logger.Fatal("line CONNECT", zap.Error(err))
	}

	openRouterClient := openrouter.NewClient(
		cfg.OpenRouterConfig.ApiKey,
	)
	kafkaProducer, err := scramkafka.NewSyncProducer(cfg.KafkaConfig)
	if err != nil {
		logger.Fatal("Fail Create NewSyncProducer", zap.Error(err))
	}
	defer func() {
		if err = kafkaProducer.Close(); err != nil {
			logger.Fatal("Fail Close SyncProducer", zap.Error(err))
		}
	}()
	logger.Info("Kafka SyncProducer Connected !!")

	metaClient, err := meta_ai.NewMetaAI("", "", nil)
	if err != nil {
		log.Fatalf("error creating MetaAI instance: %v", err)
	}
	app.Use(middleware.OTelFiberMiddleware(cfg.Server.Name))

	app.Use(middleware.AuditLogger())
	// router
	app.Post("/line/event/hook", func(c *fiber.Ctx) error {

		fmt.Println(string(c.BodyRaw()))
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	},
	)

	group := app.Group(fmt.Sprintf("/%s/api/v1", cfg.Server.Name))
	group.Get("/health", func(c *fiber.Ctx) error {
		return api.Ok(c, versionDeploy)
	})
	// ai
	group.Post("/ai/prompt", ai_proxy.NewPromptHandler(
		gemini.SendTextAndGetText(clientGemini),
		gpt.SendTextAndGetText(clientOpenAi),
		metaClient,
	))

	//TODO wait credit 10$
	group.Post("/ai/chat/complete", ai_proxy.NewChatCompletionHandler(
		dbPool,
		openRouterClient,
	))
	group.Get("/ai/model/list", ai_proxy.NewChatModelAllHandler(
		dbPool,
		cache.GetRedis(redisCMD),
		cache.SetRedis(redisCMD),
	))

	// proxy
	group.Post("/line/notification", line.NewLineNotificationHandler(
		notificationFunc,
	))
	group.Post("/email/send", email.NewEmailHandler(
		cfg.EmailConfig.Email,
		cfg.EmailConfig.User,
		cfg.EmailConfig.Pass,
	))

	// webhook api
	group.Post("/webhook", internal_webhook.NewWebHookHandler(
		kafka.NewSyncSendMessage(kafkaProducer, cfg.KafkaConfig.Topic.WebHookTopic),
		cache.SetRedis(redisCMD),
		dbPool,
	))
	// inquiry api
	group.Post("/log/inquiry", logproxy.NewLogInquiryHandler(
		cache.GetRedis(redisCMD),
		dbPool,
	))

	logger.Info(fmt.Sprintf("/%s/api/v1", cfg.Server.Name))
	if err = app.Listen(fmt.Sprintf(":%v", cfg.Server.Port)); err != nil {
		logger.Fatal(err.Error())
	}

}

func initFiber() *fiber.App {
	app := fiber.New(
		fiber.Config{
			ReadTimeout:           5 * time.Second,
			WriteTimeout:          5 * time.Second,
			IdleTimeout:           30 * time.Second,
			DisableStartupMessage: true,
			CaseSensitive:         true,
			StrictRouting:         true,
		},
	)
	app.Use(cors.New(cors.ConfigDefault))
	app.Use(SetHeaderID())
	return app
}

func SetHeaderID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		randomTrace := uuid.New().String()
		traceId := c.Get("traceId")
		reqId := c.Get("requestId")
		if traceId == "" {
			traceId = randomTrace
		}
		if reqId == "" {
			return api.BadRequest(c, "requestId is required")
		}

		c.Accepts(fiber.MIMEApplicationJSON)
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		c.Request().Header.Set("traceId", traceId)
		return c.Next()
	}
}
