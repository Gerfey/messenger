package transport_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
	"github.com/gerfey/messenger/transport"
)

type TestMessage struct {
	Text string
}

type AnotherMessage struct {
	Value int
}

func TestNewSenderLocator(t *testing.T) {
	t.Run("create new locator", func(t *testing.T) {
		locator := transport.NewSenderLocator()

		require.NotNil(t, locator)
		assert.IsType(t, &transport.SenderLocator{}, locator)

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)
		assert.Empty(t, senders)
	})
}

func TestSenderLocator_Register(t *testing.T) {
	t.Run("register single transport", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		tr := &helpers.TestTransport{}

		err := locator.Register("test-transport", tr)

		require.NoError(t, err)

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"test-transport"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, tr, senders[0])
	})

	t.Run("register multiple transports", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		err1 := locator.Register("transport1", transport1)
		err2 := locator.Register("transport2", transport2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"transport1", "transport2"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 2)
		assert.Contains(t, senders, transport1)
		assert.Contains(t, senders, transport2)
	})

	t.Run("register duplicate transport overwrites", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		err1 := locator.Register("same-name", transport1)
		err2 := locator.Register("same-name", transport2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"same-name"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, transport2, senders[0])
	})
}

func TestSenderLocator_RegisterMessageType(t *testing.T) {
	t.Run("register message type with single transport", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		transport1 := &helpers.TestTransport{}

		err := locator.Register("transport1", transport1)
		require.NoError(t, err)

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"transport1"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, transport1, senders[0])
	})

	t.Run("register message type with multiple transports", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("transport1", transport1))
		require.NoError(t, locator.Register("transport2", transport2))

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"transport1", "transport2"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 2)
		assert.Contains(t, senders, transport1)
		assert.Contains(t, senders, transport2)
	})

	t.Run("register message type with non-existing transport", func(t *testing.T) {
		locator := transport.NewSenderLocator()

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"non-existing"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		assert.Empty(t, senders)
	})
}

func TestSenderLocator_SetFallback(t *testing.T) {
	t.Run("set fallback with single transport", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		fallbackTransport := &helpers.TestTransport{}

		require.NoError(t, locator.Register("fallback", fallbackTransport))
		locator.SetFallback([]string{"fallback"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, fallbackTransport, senders[0])
	})

	t.Run("set fallback with multiple transports", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		fallback1 := &helpers.TestTransport{}
		fallback2 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("fallback1", fallback1))
		require.NoError(t, locator.Register("fallback2", fallback2))
		locator.SetFallback([]string{"fallback1", "fallback2"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 2)
		assert.Contains(t, senders, fallback1)
		assert.Contains(t, senders, fallback2)
	})
}

func TestSenderLocator_GetSenders_Priority(t *testing.T) {
	t.Run("TransportNameStamp has highest priority", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		stampTransport := &helpers.TestTransport{}
		mappedTransport := &helpers.TestTransport{}
		fallbackTransport := &helpers.TestTransport{}

		require.NoError(t, locator.Register("stamp", stampTransport))
		require.NoError(t, locator.Register("mapped", mappedTransport))
		require.NoError(t, locator.Register("fallback", fallbackTransport))

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"mapped"})
		locator.SetFallback([]string{"fallback"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		env = env.WithStamp(stamps.TransportNameStamp{Transports: []string{"stamp"}})

		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, stampTransport, senders[0])
	})

	t.Run("sendersMap has second priority", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		mappedTransport := &helpers.TestTransport{}
		fallbackTransport := &helpers.TestTransport{}

		require.NoError(t, locator.Register("mapped", mappedTransport))
		require.NoError(t, locator.Register("fallback", fallbackTransport))

		msgType := reflect.TypeOf(&TestMessage{})
		locator.RegisterMessageType(msgType, []string{"mapped"})
		locator.SetFallback([]string{"fallback"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, mappedTransport, senders[0])
	})

	t.Run("fallback has lowest priority", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		fallbackTransport := &helpers.TestTransport{}

		require.NoError(t, locator.Register("fallback", fallbackTransport))
		locator.SetFallback([]string{"fallback"})

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, fallbackTransport, senders[0])
	})
}

func TestSenderLocator_GetSenders_Deduplication(t *testing.T) {
	t.Run("prevents duplicate senders", func(t *testing.T) {
		locator := transport.NewSenderLocator()
		transport1 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("transport1", transport1))

		env := envelope.NewEnvelope(&TestMessage{Text: "test"})
		env = env.WithStamp(stamps.TransportNameStamp{
			Transports: []string{"transport1", "transport1", "transport1"},
		})

		senders := locator.GetSenders(env)

		require.Len(t, senders, 1)
		assert.Equal(t, transport1, senders[0])
	})
}

func TestSenderLocator_Integration(t *testing.T) {
	t.Run("complex scenario with all features", func(t *testing.T) {
		locator := transport.NewSenderLocator()

		amqp := &helpers.TestTransport{}
		redis := &helpers.TestTransport{}
		sync := &helpers.TestTransport{}

		require.NoError(t, locator.Register("amqp", amqp))
		require.NoError(t, locator.Register("redis", redis))
		require.NoError(t, locator.Register("sync", sync))

		testMsgType := reflect.TypeOf(&TestMessage{})
		anotherMsgType := reflect.TypeOf(&AnotherMessage{})

		locator.RegisterMessageType(testMsgType, []string{"amqp", "redis"})
		locator.RegisterMessageType(anotherMsgType, []string{"redis"})
		locator.SetFallback([]string{"sync"})

		env1 := envelope.NewEnvelope(&TestMessage{Text: "test1"})
		env1 = env1.WithStamp(stamps.TransportNameStamp{Transports: []string{"sync"}})

		senders1 := locator.GetSenders(env1)
		require.Len(t, senders1, 1)
		assert.Equal(t, sync, senders1[0])

		env2 := envelope.NewEnvelope(&TestMessage{Text: "test2"})

		senders2 := locator.GetSenders(env2)
		require.Len(t, senders2, 2)
		assert.Contains(t, senders2, amqp)
		assert.Contains(t, senders2, redis)

		type UnknownMessage struct{ Data string }
		env3 := envelope.NewEnvelope(&UnknownMessage{Data: "test3"})

		senders3 := locator.GetSenders(env3)
		require.Len(t, senders3, 1)
		assert.Equal(t, sync, senders3[0])

		env4 := envelope.NewEnvelope(&AnotherMessage{Value: 42})

		senders4 := locator.GetSenders(env4)

		require.Len(t, senders4, 1)
		assert.Equal(t, redis, senders4[0])
	})
}
