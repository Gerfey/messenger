package middleware

import (
	"context"
	"log/slog"

	"github.com/gerfey/messenger/api"
)

type LoggerMiddleware struct {
	logger *slog.Logger
}

func NewLoggerMiddleware(log *slog.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: log,
	}
}

func (l *LoggerMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	l.logger.Info("message", "message", env.Message())

	return next(ctx, env)
}
