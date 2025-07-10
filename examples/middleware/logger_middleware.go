package middleware

import (
	"context"
	"log/slog"

	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/middlewares"
)

type LoggerMiddleware struct {
	logger *slog.Logger
}

func NewLoggerMiddleware(log *slog.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: log,
	}
}

func (l *LoggerMiddleware) Handle(ctx context.Context, env *envelope.Envelope, next middlewares.NextFunc) (*envelope.Envelope, error) {
	l.logger.Info("message", "message", env.Message())

	return next(ctx, env)
}
