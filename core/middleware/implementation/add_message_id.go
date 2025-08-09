package implementation

import (
	"context"

	"github.com/google/uuid"

	"github.com/gerfey/messenger/core/envelope"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

type AddMessageIDMiddleware struct{}

func NewAddMessageIDMiddleware() *AddMessageIDMiddleware {
	return &AddMessageIDMiddleware{}
}

func (m *AddMessageIDMiddleware) Handle(
	ctx context.Context,
	env api.Envelope,
	next api.NextFunc,
) (api.Envelope, error) {
	if _, ok := envelope.LastStampOf[stamps.MessageIDStamp](env); !ok {
		env = env.WithStamp(stamps.MessageIDStamp{
			MessageID: uuid.New().String(),
		})
	}

	return next(ctx, env)
}
