package sync

import (
	"context"
	"errors"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type Transport struct {
	locator api.BusLocator
}

func NewTransport(locator api.BusLocator) api.Transport {
	return &Transport{locator: locator}
}

func (t *Transport) Send(_ context.Context, env api.Envelope) error {
	busNameStump, ok := envelope.LastStampOf[stamps.BusNameStamp](env)
	if !ok {
		return errors.New("no BusNameStamp found in envelope")
	}

	messageBus, ok := t.locator.Get(busNameStump.Name)
	if !ok {
		return errors.New("no default transport")
	}

	env = env.WithStamp(stamps.ReceivedStamp{Transport: t.Name()})

	_, err := messageBus.Dispatch(context.Background(), env)
	if err != nil {
		return err
	}

	return nil
}

func (t *Transport) Receive(_ context.Context, _ func(context.Context, api.Envelope) error) error {
	return errors.New("you cannot receive messages from the SyncTransport")
}

func (t *Transport) Name() string {
	return "sync"
}
