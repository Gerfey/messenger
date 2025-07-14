package middleware

import (
	"context"
	"log/slog"

	"github.com/gerfey/messenger/api"
)

type ExampleLoggerMiddleware struct {
	logger *slog.Logger
}

func NewExampleLoggerMiddleware(log *slog.Logger) *ExampleLoggerMiddleware {
	return &ExampleLoggerMiddleware{
		logger: log,
	}
}

func (l *ExampleLoggerMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	l.logger.Info("message", "message", env.Message())

	return next(ctx, env)
}
