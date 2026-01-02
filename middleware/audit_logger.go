package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/home-server7795544/home-server/gateway/home-proxy/internal/logz"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func AuditLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		ctx := c.UserContext()
		reqID := c.Get("requestId")

		logger := logz.WithTrace(ctx, logz.NewLogger(), reqID)

		span := trace.SpanFromContext(ctx)
		reqBody := c.Body()
		span.AddEvent("http.request", trace.WithAttributes(
			attribute.String("request.id", reqID),
			attribute.Int("body.bytes", len(reqBody)),
		))

		logger.Info("http_request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Time("start", start),
			zap.Int("req_body_bytes", len(reqBody)),
		)

		err := c.Next()

		status := c.Response().StatusCode()
		resBody := c.Response().Body()
		durationMs := time.Since(start).Milliseconds()

		// add response event to span
		span.AddEvent("http.response", trace.WithAttributes(
			attribute.Int("status", status),
			attribute.Int64("duration_ms", durationMs),
			attribute.Int("body.bytes", len(resBody)),
		))

		logger.Info("http_response",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Int64("duration_ms", durationMs),
			zap.Int("res_body_bytes", len(resBody)),
		)

		return err
	}
}
