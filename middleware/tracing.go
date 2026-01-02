package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// --- fasthttp header carrier for Extract ---
type reqCarrier struct{ h *fasthttp.RequestHeader }

func (c reqCarrier) Get(key string) string { return string(c.h.Peek(key)) }
func (c reqCarrier) Set(key, val string)   { c.h.Set(key, val) }
func (c reqCarrier) Keys() []string        { return nil }

func OTelFiberMiddleware(serviceName string) fiber.Handler {
	tr := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		savedCtx := c.UserContext()
		ctx := otel.GetTextMapPropagator().Extract(savedCtx, reqCarrier{h: &c.Context().Request.Header})

		spanName := spanNameFormatterFiber(c)
		ctx, span := tr.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		start := time.Now()
		defer span.End()

		c.SetUserContext(ctx)

		span.SetAttributes(
			attribute.String("service.name", serviceName),

			attribute.String("http.method", c.Method()),
			attribute.String("http.route", c.OriginalURL()),
			attribute.String("http.target", c.OriginalURL()),
			attribute.String("http.scheme", string(c.Protocol())),
			attribute.String("net.peer.ip", c.IP()),
			attribute.String("user_agent", c.Get("User-Agent")),
		)

		err := c.Next()

		status := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Int64("http.duration_ms", time.Since(start).Milliseconds()),
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else if status >= 500 {
			span.SetStatus(codes.Error, "server_error")
		} else if status >= 400 {
			span.SetStatus(codes.Error, "client_error")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

func spanNameFormatterFiber(c *fiber.Ctx) string {
	method := strings.ToUpper(c.Method())
	path := c.Route().Path
	if path == "" {
		path = c.Path()
	}
	return method + " " + path
}
