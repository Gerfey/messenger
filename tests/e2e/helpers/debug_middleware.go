package helpers

import (
	"context"
	"log/slog"

	"github.com/gerfey/messenger/api"
)

type DebugMiddleware struct {
	name   string
	logger *slog.Logger
}

func NewDebugMiddleware(name string, logger *slog.Logger) *DebugMiddleware {
	return &DebugMiddleware{
		name:   name,
		logger: logger,
	}
}

func (d *DebugMiddleware) Handle(ctx context.Context, env api.Envelope, next api.NextFunc) (api.Envelope, error) {
	d.logger.Info("DEBUG MIDDLEWARE BEFORE")
	
	result, err := next(ctx, env)
	
	d.logger.Info("DEBUG MIDDLEWARE AFTER")
	
	return result, err
}
