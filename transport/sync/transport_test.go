package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/mocks"
	"github.com/gerfey/messenger/transport/sync"
)

func TestNewTransport(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)

	transport := sync.NewTransport(mockLocator)

	assert.NotNil(t, transport)
	assert.IsType(t, &sync.Transport{}, transport)
}

func TestTransport_Send_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	mockBus := mocks.NewMockMessageBus(ctrl)
	transport := sync.NewTransport(mockLocator)

	msg := &testMessage{content: "test message"}
	env := envelope.NewEnvelope(msg)
	busName := "test-bus"
	env = env.WithStamp(stamps.BusNameStamp{Name: busName})

	mockLocator.EXPECT().Get(busName).Return(mockBus, true)

	mockBus.EXPECT().
		Dispatch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(env, nil)

	err := transport.Send(t.Context(), env)

	require.NoError(t, err)
}

func TestTransport_Send_NoBusNameStamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	transport := sync.NewTransport(mockLocator)

	msg := &testMessage{content: "test message"}
	env := envelope.NewEnvelope(msg)

	err := transport.Send(t.Context(), env)

	require.Error(t, err)
	assert.Equal(t, "no BusNameStamp found in envelope", err.Error())
}

func TestTransport_Send_NoBusFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	transport := sync.NewTransport(mockLocator)

	msg := &testMessage{content: "test message"}
	env := envelope.NewEnvelope(msg)
	busName := "non-existent-bus"
	env = env.WithStamp(stamps.BusNameStamp{Name: busName})

	mockLocator.EXPECT().Get(busName).Return(nil, false)

	err := transport.Send(t.Context(), env)

	require.Error(t, err)
	assert.Equal(t, "no default transport", err.Error())
}

func TestTransport_Send_DispatchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	mockBus := mocks.NewMockMessageBus(ctrl)
	transport := sync.NewTransport(mockLocator)

	msg := &testMessage{content: "test message"}
	env := envelope.NewEnvelope(msg)
	busName := "test-bus"
	env = env.WithStamp(stamps.BusNameStamp{Name: busName})

	dispatchErr := errors.New("dispatch error")

	mockLocator.EXPECT().Get(busName).Return(mockBus, true)
	mockBus.EXPECT().Dispatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, dispatchErr)

	err := transport.Send(t.Context(), env)

	require.Error(t, err)
	assert.Equal(t, dispatchErr, err)
}

func TestTransport_Receive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	transport := sync.NewTransport(mockLocator)

	handler := func(context.Context, api.Envelope) error { return nil }

	err := transport.Receive(t.Context(), handler)

	require.Error(t, err)
	assert.Equal(t, "you cannot receive messages from the SyncTransport", err.Error())
}

func TestTransport_Name(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLocator := mocks.NewMockBusLocator(ctrl)
	transport := sync.NewTransport(mockLocator)

	name := transport.Name()

	assert.Equal(t, "sync", name)
}

type testMessage struct {
	content string
}
