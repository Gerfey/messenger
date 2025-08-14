package builder_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/builder"

	"github.com/gerfey/messenger/core/stamps"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewStaticTypeResolver(t *testing.T) {
	t.Run("create new resolver", func(t *testing.T) {
		resolver := builder.NewResolver()

		require.NotNil(t, resolver)

		_, err := resolver.ResolveMessageType("non.existing.Type")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown message type")
	})
}

func TestResolver_Register(t *testing.T) {
	t.Run("register and resolve type by string", func(t *testing.T) {
		resolver := builder.NewResolver()
		msgType := reflect.TypeOf(&helpers.TestMessage{})
		typeStr := "custom.TestMessage"

		resolver.Register(typeStr, msgType)

		resolvedType, err := resolver.ResolveMessageType(typeStr)
		require.NoError(t, err)
		assert.Equal(t, msgType, resolvedType)
	})

	t.Run("register multiple types", func(t *testing.T) {
		resolver := builder.NewResolver()

		msgType1 := reflect.TypeOf(&helpers.TestMessage{})
		msgType2 := reflect.TypeOf(&helpers.ComplexMessage{})

		resolver.Register("type1", msgType1)
		resolver.Register("type2", msgType2)

		resolvedType1, err1 := resolver.ResolveMessageType("type1")
		resolvedType2, err2 := resolver.ResolveMessageType("type2")

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, msgType1, resolvedType1)
		assert.Equal(t, msgType2, resolvedType2)
	})

	t.Run("register same type string overwrites", func(t *testing.T) {
		resolver := builder.NewResolver()

		msgType1 := reflect.TypeOf(&helpers.TestMessage{})
		msgType2 := reflect.TypeOf(&helpers.ComplexMessage{})

		resolver.Register("sametype", msgType1)
		resolver.Register("sametype", msgType2)

		resolvedType, err := resolver.ResolveMessageType("sametype")
		require.NoError(t, err)
		assert.Equal(t, msgType2, resolvedType)
	})
}

func TestResolver_RegisterMessage(t *testing.T) {
	t.Run("register message by instance", func(t *testing.T) {
		resolver := builder.NewResolver()
		msg := &helpers.TestMessage{ID: "test", Content: "content"}

		resolver.RegisterMessage(msg)

		expectedType := reflect.TypeOf(msg)
		expectedKey := expectedType.String()

		resolvedType, err := resolver.ResolveMessageType(expectedKey)
		require.NoError(t, err)
		assert.Equal(t, expectedType, resolvedType)
	})

	t.Run("register multiple different messages", func(t *testing.T) {
		resolver := builder.NewResolver()

		msg1 := &helpers.TestMessage{}
		msg2 := &helpers.ComplexMessage{}
		msg3 := helpers.SimpleMessage("test")

		resolver.RegisterMessage(msg1)
		resolver.RegisterMessage(msg2)
		resolver.RegisterMessage(msg3)

		type1 := reflect.TypeOf(msg1)
		type2 := reflect.TypeOf(msg2)
		type3 := reflect.TypeOf(msg3)

		resolvedType1, err1 := resolver.ResolveMessageType(type1.String())
		resolvedType2, err2 := resolver.ResolveMessageType(type2.String())
		resolvedType3, err3 := resolver.ResolveMessageType(type3.String())

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		assert.Equal(t, type1, resolvedType1)
		assert.Equal(t, type2, resolvedType2)
		assert.Equal(t, type3, resolvedType3)
	})

	t.Run("register same message type overwrites", func(t *testing.T) {
		resolver := builder.NewResolver()

		msg1 := &helpers.TestMessage{ID: "1"}
		msg2 := &helpers.TestMessage{ID: "2"}

		resolver.RegisterMessage(msg1)
		resolver.RegisterMessage(msg2)

		expectedType := reflect.TypeOf(msg1)
		resolvedType, err := resolver.ResolveMessageType(expectedType.String())
		require.NoError(t, err)
		assert.Equal(t, expectedType, resolvedType)
	})

	t.Run("register nil message", func(t *testing.T) {
		resolver := builder.NewResolver()

		require.Panics(t, func() {
			resolver.RegisterMessage(nil)
		})
	})
}

func TestResolver_RegisterStamp(t *testing.T) {
	t.Run("register stamp by instance", func(t *testing.T) {
		resolver := builder.NewResolver()
		stamp := &stamps.BusNameStamp{Name: "test"}

		resolver.RegisterStamp(stamp)

		expectedType := reflect.TypeOf(stamp)
		expectedKey := expectedType.String()

		resolvedType, err := resolver.ResolveStampType(expectedKey)
		require.NoError(t, err)
		assert.Equal(t, expectedType, resolvedType)
	})

	t.Run("register multiple different stamps", func(t *testing.T) {
		resolver := builder.NewResolver()

		stamp1 := &stamps.BusNameStamp{}
		stamp2 := &stamps.SentStamp{}
		stamp3 := &helpers.TestStamp{}

		resolver.RegisterStamp(stamp1)
		resolver.RegisterStamp(stamp2)
		resolver.RegisterStamp(stamp3)

		type1 := reflect.TypeOf(stamp1)
		type2 := reflect.TypeOf(stamp2)
		type3 := reflect.TypeOf(stamp3)

		resolvedType1, err1 := resolver.ResolveStampType(type1.String())
		resolvedType2, err2 := resolver.ResolveStampType(type2.String())
		resolvedType3, err3 := resolver.ResolveStampType(type3.String())

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		assert.Equal(t, type1, resolvedType1)
		assert.Equal(t, type2, resolvedType2)
		assert.Equal(t, type3, resolvedType3)
	})
}

func TestResolver_ResolveMessageType(t *testing.T) {
	t.Run("resolve existing message type", func(t *testing.T) {
		resolver := builder.NewResolver()
		msg := &helpers.TestMessage{}

		resolver.RegisterMessage(msg)

		expectedType := reflect.TypeOf(msg)
		resolvedType, err := resolver.ResolveMessageType(expectedType.String())

		require.NoError(t, err)
		assert.Equal(t, expectedType, resolvedType)
	})

	t.Run("resolve non-existing message type", func(t *testing.T) {
		resolver := builder.NewResolver()

		resolvedType, err := resolver.ResolveMessageType("non.existing.Type")

		require.Error(t, err)
		assert.Nil(t, resolvedType)
		assert.Contains(t, err.Error(), "unknown message type: non.existing.Type")
	})

	t.Run("resolve multiple registered types", func(t *testing.T) {
		resolver := builder.NewResolver()

		msg1 := &helpers.TestMessage{}
		msg2 := &helpers.ComplexMessage{}

		resolver.RegisterMessage(msg1)
		resolver.RegisterMessage(msg2)

		type1 := reflect.TypeOf(msg1)
		type2 := reflect.TypeOf(msg2)

		resolvedType1, err1 := resolver.ResolveMessageType(type1.String())
		resolvedType2, err2 := resolver.ResolveMessageType(type2.String())

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, type1, resolvedType1)
		assert.Equal(t, type2, resolvedType2)
	})

	t.Run("resolve with empty type name", func(t *testing.T) {
		resolver := builder.NewResolver()

		resolvedType, err := resolver.ResolveMessageType("")

		require.Error(t, err)
		assert.Nil(t, resolvedType)
		assert.Contains(t, err.Error(), "unknown message type: ")
	})
}

func TestResolver_ResolveStampType(t *testing.T) {
	t.Run("resolve existing stamp type", func(t *testing.T) {
		resolver := builder.NewResolver()
		stamp := &stamps.BusNameStamp{}

		resolver.RegisterStamp(stamp)

		expectedType := reflect.TypeOf(stamp)
		resolvedType, err := resolver.ResolveStampType(expectedType.String())

		require.NoError(t, err)
		assert.Equal(t, expectedType, resolvedType)
	})

	t.Run("resolve non-existing stamp type", func(t *testing.T) {
		resolver := builder.NewResolver()

		resolvedType, err := resolver.ResolveStampType("non.existing.StampType")

		require.Error(t, err)
		assert.Nil(t, resolvedType)
		assert.Contains(t, err.Error(), "unknown stamp type: non.existing.StampType")
	})

	t.Run("resolve multiple registered stamp types", func(t *testing.T) {
		resolver := builder.NewResolver()

		stamp1 := &stamps.BusNameStamp{}
		stamp2 := &stamps.SentStamp{}
		stamp3 := &helpers.TestStamp{}

		resolver.RegisterStamp(stamp1)
		resolver.RegisterStamp(stamp2)
		resolver.RegisterStamp(stamp3)

		type1 := reflect.TypeOf(stamp1)
		type2 := reflect.TypeOf(stamp2)
		type3 := reflect.TypeOf(stamp3)

		resolvedType1, err1 := resolver.ResolveStampType(type1.String())
		resolvedType2, err2 := resolver.ResolveStampType(type2.String())
		resolvedType3, err3 := resolver.ResolveStampType(type3.String())

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		assert.Equal(t, type1, resolvedType1)
		assert.Equal(t, type2, resolvedType2)
		assert.Equal(t, type3, resolvedType3)
	})
}

func TestResolver_Integration(t *testing.T) {
	t.Run("full workflow with messages and stamps", func(t *testing.T) {
		resolver := builder.NewResolver()

		msg1 := &helpers.TestMessage{ID: "test"}
		msg2 := &helpers.ComplexMessage{ID: "complex"}

		resolver.RegisterMessage(msg1)
		resolver.RegisterMessage(msg2)

		stamp1 := &stamps.BusNameStamp{Name: "bus"}
		stamp2 := &helpers.TestStamp{Value: "test"}

		resolver.RegisterStamp(stamp1)
		resolver.RegisterStamp(stamp2)

		msgType1 := reflect.TypeOf(msg1)
		msgType2 := reflect.TypeOf(msg2)
		stampType1 := reflect.TypeOf(stamp1)
		stampType2 := reflect.TypeOf(stamp2)

		resolvedMsg1, err1 := resolver.ResolveMessageType(msgType1.String())
		resolvedMsg2, err2 := resolver.ResolveMessageType(msgType2.String())

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, msgType1, resolvedMsg1)
		assert.Equal(t, msgType2, resolvedMsg2)

		resolvedStamp1, err3 := resolver.ResolveStampType(stampType1.String())
		resolvedStamp2, err4 := resolver.ResolveStampType(stampType2.String())

		require.NoError(t, err3)
		require.NoError(t, err4)
		assert.Equal(t, stampType1, resolvedStamp1)
		assert.Equal(t, stampType2, resolvedStamp2)

		_, err5 := resolver.ResolveMessageType("non.existing.Message")
		_, err6 := resolver.ResolveStampType("non.existing.Stamp")

		require.Error(t, err5)
		require.Error(t, err6)
	})
}
