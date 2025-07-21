package handler_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/handler"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewHandlerLocator(t *testing.T) {
	locator := handler.NewHandlerLocator()

	assert.NotNil(t, locator)
	assert.Empty(t, locator.GetAll())
}

func TestLocator_Register_ValidHandlers(t *testing.T) {
	t.Run("register valid handler", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandler{}

		err := locator.Register(validHandler)

		require.NoError(t, err)

		handlers := locator.GetAll()
		assert.Len(t, handlers, 1)
		assert.Equal(t, reflect.TypeOf(&helpers.TestMessage{}), handlers[0].InputType)
	})

	t.Run("register handler with result", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandlerWithResult{}

		err := locator.Register(validHandler)

		require.NoError(t, err)
		handlers := locator.GetAll()
		assert.Len(t, handlers, 1)
	})

	t.Run("register handler with bus name", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandlerWithBusName{}

		err := locator.Register(validHandler)

		require.NoError(t, err)
		handlers := locator.GetAll()
		assert.Len(t, handlers, 1)
		assert.Equal(t, "test-bus", handlers[0].BusName)
	})

	t.Run("register multiple handlers for same message type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.ValidHandlerWithResult{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		msg := &helpers.TestMessage{}
		handlers := locator.Get(msg)
		assert.Len(t, handlers, 2)
	})

	t.Run("register handlers for different message types", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.AnotherValidHandler{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		allHandlers := locator.GetAll()
		assert.Len(t, allHandlers, 2)

		testMsgHandlers := locator.Get(&helpers.TestMessage{})
		assert.Len(t, testMsgHandlers, 1)

		complexMsgHandlers := locator.Get(&helpers.ComplexMessage{})
		assert.Len(t, complexMsgHandlers, 1)
	})
}

func TestLocator_Register_InvalidHandlers(t *testing.T) {
	t.Run("handler without Handle method", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerNoMethod{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not have a Handle method")
	})

	t.Run("handler with wrong number of parameters", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerWrongParams{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Handle method must accept exactly 2 parameters")
	})

	t.Run("handler with wrong first parameter type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerWrongFirstParam{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "first parameter must be context.Context")
	})

	t.Run("handler with too many parameters", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerTooManyParams{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Handle method must accept exactly 2 parameters")
	})

	t.Run("handler with no return values", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerNoReturn{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Handle method must return error or (result, error)")
	})

	t.Run("handler with wrong return type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerWrongReturn{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "last return value must be error")
	})

	t.Run("handler with too many return values", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		invalidHandler := &helpers.InvalidHandlerTooManyReturns{}

		err := locator.Register(invalidHandler)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Handle method must return error or (result, error)")
	})
}

func TestLocator_Get(t *testing.T) {
	t.Run("get handlers for registered message type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandler{}
		err := locator.Register(validHandler)
		require.NoError(t, err)

		msg := &helpers.TestMessage{}
		handlers := locator.Get(msg)

		assert.Len(t, handlers, 1)
		assert.Equal(t, reflect.TypeOf(msg), handlers[0].InputType)
	})

	t.Run("get handlers for unregistered message type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()

		msg := &helpers.TestMessage{}
		handlers := locator.Get(msg)

		assert.Empty(t, handlers)
		assert.NotNil(t, handlers)
	})

	t.Run("get multiple handlers for same message type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.ValidHandlerWithResult{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		msg := &helpers.TestMessage{}
		handlers := locator.Get(msg)

		assert.Len(t, handlers, 2)
	})
}

func TestLocator_GetAll(t *testing.T) {
	t.Run("get all handlers from empty locator", func(t *testing.T) {
		locator := handler.NewHandlerLocator()

		handlers := locator.GetAll()

		assert.Empty(t, handlers)
		assert.NotNil(t, handlers)
	})

	t.Run("get all handlers with multiple registrations", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.AnotherValidHandler{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		handlers := locator.GetAll()

		assert.Len(t, handlers, 2)

		inputTypes := make(map[reflect.Type]bool)
		for _, h := range handlers {
			inputTypes[h.InputType] = true
		}
		assert.True(t, inputTypes[reflect.TypeOf(&helpers.TestMessage{})])
		assert.True(t, inputTypes[reflect.TypeOf(&helpers.ComplexMessage{})])
	})
}

func TestLocator_ResolveMessageType(t *testing.T) {
	t.Run("resolve registered message type by string", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandler{}
		err := locator.Register(validHandler)
		require.NoError(t, err)

		msgType, err := locator.ResolveMessageType("*helpers.TestMessage")

		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(&helpers.TestMessage{}), msgType)
	})

	t.Run("resolve registered message type by element string", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandler{}
		err := locator.Register(validHandler)
		require.NoError(t, err)

		msgType, err := locator.ResolveMessageType("helpers.TestMessage")

		require.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(&helpers.TestMessage{}), msgType)
	})

	t.Run("resolve unregistered message type", func(t *testing.T) {
		locator := handler.NewHandlerLocator()

		msgType, err := locator.ResolveMessageType("UnknownMessage")

		require.Error(t, err)
		assert.Nil(t, msgType)
		assert.Contains(t, err.Error(), "message type \"UnknownMessage\" not found in registry")
	})

	t.Run("resolve multiple registered types", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.AnotherValidHandler{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		msgType1, err1 := locator.ResolveMessageType("*helpers.TestMessage")
		msgType2, err2 := locator.ResolveMessageType("*helpers.ComplexMessage")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, reflect.TypeOf(&helpers.TestMessage{}), msgType1)
		assert.Equal(t, reflect.TypeOf(&helpers.ComplexMessage{}), msgType2)
	})
}

func TestLocator_HandlerStr_RuntimeFuncName(t *testing.T) {
	t.Run("struct handler string representation", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandler{}

		err := locator.Register(validHandler)
		require.NoError(t, err)

		handlers := locator.GetAll()
		assert.Len(t, handlers, 1)

		handlerStr := handlers[0].HandlerStr
		assert.NotContains(t, handlerStr, "invalid")
		assert.NotContains(t, handlerStr, "no pointer")
		assert.NotContains(t, handlerStr, "no symbol")
		assert.NotEmpty(t, handlerStr)
	})

	t.Run("handler with bus name string representation", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		validHandler := &helpers.ValidHandlerWithBusName{}

		err := locator.Register(validHandler)
		require.NoError(t, err)

		handlers := locator.GetAll()
		assert.Len(t, handlers, 1)

		handlerStr := handlers[0].HandlerStr
		assert.NotContains(t, handlerStr, "invalid")
		assert.NotEmpty(t, handlerStr)

		assert.Equal(t, "test-bus", handlers[0].BusName)
	})

	t.Run("multiple handlers string representation", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		handler1 := &helpers.ValidHandler{}
		handler2 := &helpers.ValidHandlerWithResult{}

		err1 := locator.Register(handler1)
		err2 := locator.Register(handler2)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		handlers := locator.GetAll()
		assert.Len(t, handlers, 2)

		for _, h := range handlers {
			assert.NotContains(t, h.HandlerStr, "invalid")
			assert.NotContains(t, h.HandlerStr, "no pointer")
			assert.NotContains(t, h.HandlerStr, "no symbol")
			assert.NotEmpty(t, h.HandlerStr)
		}
	})

	t.Run("handler string contains type information", func(t *testing.T) {
		locator := handler.NewHandlerLocator()
		testHandler := &helpers.ValidHandler{}
		complexHandler := &helpers.AnotherValidHandler{}

		err1 := locator.Register(testHandler)
		err2 := locator.Register(complexHandler)
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		allHandlers := locator.GetAll()
		assert.Len(t, allHandlers, 2)

		testMsgHandlers := locator.Get(&helpers.TestMessage{})
		complexMsgHandlers := locator.Get(&helpers.ComplexMessage{})

		assert.Len(t, testMsgHandlers, 1)
		assert.Len(t, complexMsgHandlers, 1)

		assert.NotEmpty(t, testMsgHandlers[0].HandlerStr)
		assert.NotEmpty(t, complexMsgHandlers[0].HandlerStr)
		assert.NotContains(t, testMsgHandlers[0].HandlerStr, "invalid")
		assert.NotContains(t, complexMsgHandlers[0].HandlerStr, "invalid")
	})
}
