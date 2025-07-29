package inmemory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/transport/inmemory"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewTransport(t *testing.T) {
	t.Run("create transport with config", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		require.NotNil(t, transport)
		assert.IsType(t, &inmemory.Transport{}, transport)
		assert.Equal(t, "test-transport", transport.Name())
	})

	t.Run("create transport with empty config", func(t *testing.T) {
		transport := inmemory.NewTransport("in-memory")

		require.NotNil(t, transport)
		assert.Empty(t, transport.Name())
	})
}

func TestTransport_Name(t *testing.T) {
	t.Run("get transport name", func(t *testing.T) {
		transport := inmemory.NewTransport("my-transport")

		name := transport.Name()
		assert.Equal(t, "my-transport", name)
	})

	t.Run("get empty transport name", func(t *testing.T) {
		transport := inmemory.NewTransport("")

		name := transport.Name()
		assert.Empty(t, name)
	})
}

func TestTransport_Send(t *testing.T) {
	t.Run("send single message", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport").(*inmemory.Transport)

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		err := transport.Send(t.Context(), env)

		require.NoError(t, err)
	})

	t.Run("send multiple messages", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport").(*inmemory.Transport)

		msg1 := &helpers.TestMessage{Content: "test1"}
		msg2 := &helpers.TestMessage{Content: "test2"}
		msg3 := &helpers.TestMessage{Content: "test3"}

		env1 := envelope.NewEnvelope(msg1)
		env2 := envelope.NewEnvelope(msg2)
		env3 := envelope.NewEnvelope(msg3)

		err1 := transport.Send(t.Context(), env1)
		err2 := transport.Send(t.Context(), env2)
		err3 := transport.Send(t.Context(), env3)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
	})

	t.Run("send with cancelled context", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := transport.Send(ctx, env)
		require.NoError(t, err)
	})
}

func TestTransport_Receive(t *testing.T) {
	t.Run("receive single message", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		err := transport.Send(t.Context(), env)
		require.NoError(t, err)

		var receivedEnv api.Envelope
		handler := func(_ context.Context, env api.Envelope) error {
			receivedEnv = env

			return nil
		}

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err = transport.Receive(ctx, handler)

		assert.Equal(t, context.DeadlineExceeded, err)

		require.NotNil(t, receivedEnv)
		assert.Equal(t, msg, receivedEnv.Message())

		receivedStamps := envelope.StampsOf[stamps.ReceivedStamp](receivedEnv)
		require.Len(t, receivedStamps, 1)
		assert.Equal(t, "test-transport", receivedStamps[0].Transport)
	})

	t.Run("receive multiple messages", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		msg1 := &helpers.TestMessage{Content: "test1"}
		msg2 := &helpers.TestMessage{Content: "test2"}

		env1 := envelope.NewEnvelope(msg1)
		env2 := envelope.NewEnvelope(msg2)

		require.NoError(t, transport.Send(t.Context(), env1))
		require.NoError(t, transport.Send(t.Context(), env2))

		var receivedEnvs []api.Envelope
		handler := func(_ context.Context, env api.Envelope) error {
			receivedEnvs = append(receivedEnvs, env)

			return nil
		}

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err := transport.Receive(ctx, handler)
		assert.Equal(t, context.DeadlineExceeded, err)

		require.Len(t, receivedEnvs, 2)
		assert.Equal(t, msg1, receivedEnvs[0].Message())
		assert.Equal(t, msg2, receivedEnvs[1].Message())
	})

	t.Run("receive with handler error", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		msg := &helpers.TestMessage{Content: "test"}
		env := envelope.NewEnvelope(msg)

		require.NoError(t, transport.Send(t.Context(), env))

		expectedError := errors.New("handler error")
		handler := func(_ context.Context, _ api.Envelope) error {
			return expectedError
		}

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err := transport.Receive(ctx, handler)

		assert.Equal(t, expectedError, err)
	})

	t.Run("receive with cancelled context", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		handler := func(_ context.Context, _ api.Envelope) error {
			return nil
		}

		err := transport.Receive(ctx, handler)

		assert.Equal(t, context.Canceled, err)
	})

	t.Run("receive with empty queue waits", func(t *testing.T) {
		transport := inmemory.NewTransport("test-transport")

		handler := func(_ context.Context, _ api.Envelope) error {
			return nil
		}

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := transport.Receive(ctx, handler)
		duration := time.Since(start)

		assert.Equal(t, context.DeadlineExceeded, err)
		assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
	})
}

func TestTransport_Integration(t *testing.T) {
	t.Run("full send and receive workflow", func(t *testing.T) {
		transport := inmemory.NewTransport("integration-transport")

		messages := []*helpers.TestMessage{
			{Content: "message1"},
			{Content: "message2"},
			{Content: "message3"},
		}

		for _, msg := range messages {
			env := envelope.NewEnvelope(msg)
			err := transport.Send(t.Context(), env)
			require.NoError(t, err)
		}

		var receivedMessages []string
		handler := func(_ context.Context, env api.Envelope) error {
			msg := env.Message().(*helpers.TestMessage)
			receivedMessages = append(receivedMessages, msg.Content)

			receivedStamps := envelope.StampsOf[stamps.ReceivedStamp](env)
			require.Len(t, receivedStamps, 1)
			assert.Equal(t, "integration-transport", receivedStamps[0].Transport)

			return nil
		}

		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		err := transport.Receive(ctx, handler)
		assert.Equal(t, context.DeadlineExceeded, err)

		require.Len(t, receivedMessages, 3)
		assert.Equal(t, "message1", receivedMessages[0])
		assert.Equal(t, "message2", receivedMessages[1])
		assert.Equal(t, "message3", receivedMessages[2])
	})

	t.Run("concurrent send and receive", func(t *testing.T) {
		transport := inmemory.NewTransport("concurrent-transport")

		go func() {
			for range 5 {
				msg := &helpers.TestMessage{Content: "concurrent"}
				env := envelope.NewEnvelope(msg)
				_ = transport.Send(t.Context(), env)
				time.Sleep(10 * time.Millisecond)
			}
		}()

		var receivedCount int
		handler := func(_ context.Context, _ api.Envelope) error {
			receivedCount++

			return nil
		}

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()

		err := transport.Receive(ctx, handler)
		assert.Equal(t, context.DeadlineExceeded, err)

		assert.Positive(t, receivedCount)
		assert.LessOrEqual(t, receivedCount, 5)
	})
}
