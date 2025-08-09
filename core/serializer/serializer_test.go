package serializer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/builder"

	"github.com/gerfey/messenger/core/serializer"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewSerializer(t *testing.T) {
	resolver := builder.NewResolver()
	s := serializer.NewSerializer(resolver)

	require.NotNil(t, s)
}

func TestSerializer_Marshal(t *testing.T) {
	resolver := builder.NewResolver()
	s := serializer.NewSerializer(resolver)

	t.Run("marshal simple message", func(t *testing.T) {
		msg := &helpers.TestMessage{ID: "123", Content: "test message"}
		env := envelope.NewEnvelope(msg)

		body, headers, err := s.Marshal(env)

		require.NoError(t, err)
		require.NotNil(t, body)
		require.NotNil(t, headers)

		assert.Contains(t, string(body), "123")
		assert.Contains(t, string(body), "test message")
		assert.Equal(t, "*helpers.TestMessage", headers["type"])
	})

	t.Run("marshal message with stamps", func(t *testing.T) {
		msg := &helpers.TestMessage{ID: "456", Content: "message with stamps"}
		stamp := &helpers.TestStamp{Value: "test-stamp"}
		env := envelope.NewEnvelope(msg).WithStamp(stamp)

		body, headers, err := s.Marshal(env)

		require.NoError(t, err)
		require.NotNil(t, body)
		require.NotNil(t, headers)

		assert.Contains(t, string(body), "456")
		assert.Contains(t, string(body), "message with stamps")
		assert.Equal(t, "*helpers.TestMessage", headers["type"])
		assert.Contains(t, headers, "stamps")
		assert.Contains(t, headers["stamps"], "test-stamp")
	})

	t.Run("marshal error on invalid message", func(t *testing.T) {
		type BadMessage struct {
			BadField chan int
		}
		msg := &BadMessage{BadField: make(chan int)}
		env := envelope.NewEnvelope(msg)

		body, headers, err := s.Marshal(env)

		require.Error(t, err)
		assert.Nil(t, body)
		assert.Nil(t, headers)
	})
}

func TestSerializer_Unmarshal(t *testing.T) {
	resolver := builder.NewResolver()

	resolver.RegisterMessage(&helpers.TestMessage{})

	s := serializer.NewSerializer(resolver)

	t.Run("unmarshal simple message", func(t *testing.T) {
		body := []byte(`{"ID":"123","Content":"test message"}`)
		headers := map[string]string{
			"type": "*helpers.TestMessage",
		}

		env, err := s.Unmarshal(body, headers)

		require.NoError(t, err)
		require.NotNil(t, env)

		msg, ok := env.Message().(*helpers.TestMessage)
		require.True(t, ok)
		assert.Equal(t, "123", msg.ID)
		assert.Equal(t, "test message", msg.Content)
	})

	t.Run("unmarshal error missing type header", func(t *testing.T) {
		body := []byte(`{"ID":"123","Content":"test"}`)
		headers := map[string]string{}

		env, err := s.Unmarshal(body, headers)

		require.Error(t, err)
		assert.Nil(t, env)
		assert.Contains(t, err.Error(), "missing 'type' header")
	})

	t.Run("unmarshal error unknown message type", func(t *testing.T) {
		body := []byte(`{"ID":"123","Content":"test"}`)
		headers := map[string]string{
			"type": "*UnknownMessage",
		}

		env, err := s.Unmarshal(body, headers)

		require.Error(t, err)
		assert.Nil(t, env)
		assert.Contains(t, err.Error(), "unknown message type")
	})

	t.Run("unmarshal error invalid JSON body", func(t *testing.T) {
		body := []byte(`invalid json`)
		headers := map[string]string{
			"type": "*helpers.TestMessage",
		}

		env, err := s.Unmarshal(body, headers)

		require.Error(t, err)
		assert.Nil(t, env)
	})

	t.Run("unmarshal with invalid stamps JSON", func(t *testing.T) {
		body := []byte(`{"ID":"123","Content":"test"}`)
		headers := map[string]string{
			"type":   "*helpers.TestMessage",
			"stamps": "invalid json",
		}

		env, err := s.Unmarshal(body, headers)

		require.NoError(t, err)
		require.NotNil(t, env)

		msg, ok := env.Message().(*helpers.TestMessage)
		require.True(t, ok)
		assert.Equal(t, "123", msg.ID)

		stamps := env.Stamps()
		assert.Empty(t, stamps)
	})
}

func TestSerializer_Integration(t *testing.T) {
	resolver := builder.NewResolver()

	resolver.RegisterMessage(&helpers.TestMessage{})

	s := serializer.NewSerializer(resolver)

	t.Run("marshal and unmarshal roundtrip", func(t *testing.T) {
		originalMsg := &helpers.TestMessage{ID: "roundtrip", Content: "test roundtrip"}
		originalEnv := envelope.NewEnvelope(originalMsg)

		body, headers, err := s.Marshal(originalEnv)
		require.NoError(t, err)

		restoredEnv, err := s.Unmarshal(body, headers)
		require.NoError(t, err)

		restoredMsg, ok := restoredEnv.Message().(*helpers.TestMessage)
		require.True(t, ok)
		assert.Equal(t, originalMsg.ID, restoredMsg.ID)
		assert.Equal(t, originalMsg.Content, restoredMsg.Content)
	})
}
