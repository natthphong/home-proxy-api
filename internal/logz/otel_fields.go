package logz

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func WithTrace(ctx context.Context, l *zap.Logger, requestID string) *zap.Logger {
	if l == nil {
		l = zap.L()
	}
	fields := []zap.Field{}
	if requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		fields = append(fields,
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		)
	}

	return l.With(fields...)
}
