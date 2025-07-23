package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/middleware"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewMiddlewareLocator(t *testing.T) {
	t.Run("create new middleware locator", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		require.NotNil(t, locator)
	})
}

func TestLocator_Register(t *testing.T) {
	t.Run("register middleware", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		testMiddleware := &helpers.TestMiddleware{}

		locator.Register("test_middleware", testMiddleware)

		mw, err := locator.Get("test_middleware")
		require.NoError(t, err)
		assert.Equal(t, testMiddleware, mw)
	})

	t.Run("register multiple middleware", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		testMiddleware1 := &helpers.TestMiddleware{}
		testMiddleware2 := &helpers.ErrorMiddleware{}

		locator.Register("test_middleware1", testMiddleware1)
		locator.Register("test_middleware2", testMiddleware2)

		mw1, err := locator.Get("test_middleware1")
		require.NoError(t, err)
		assert.Equal(t, testMiddleware1, mw1)

		mw2, err := locator.Get("test_middleware2")
		require.NoError(t, err)
		assert.Equal(t, testMiddleware2, mw2)
	})

	t.Run("register middleware with same name overwrites", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		testMiddleware1 := &helpers.TestMiddleware{}
		testMiddleware2 := &helpers.ErrorMiddleware{}

		locator.Register("test_middleware", testMiddleware1)
		locator.Register("test_middleware", testMiddleware2)

		mw, err := locator.Get("test_middleware")
		require.NoError(t, err)
		assert.Equal(t, testMiddleware2, mw)
	})
}

func TestLocator_GetAll(t *testing.T) {
	t.Run("get all middleware from empty locator", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()

		all := locator.GetAll()
		assert.Empty(t, all)
	})

	t.Run("get all middleware", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		testMiddleware1 := &helpers.TestMiddleware{}
		testMiddleware2 := &helpers.ErrorMiddleware{}

		locator.Register("test_middleware1", testMiddleware1)
		locator.Register("test_middleware2", testMiddleware2)

		all := locator.GetAll()

		assert.Len(t, all, 2)
		assert.Contains(t, all, testMiddleware1)
		assert.Contains(t, all, testMiddleware2)
	})
}

func TestLocator_Get(t *testing.T) {
	t.Run("get existing middleware", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()
		testMiddleware := &helpers.TestMiddleware{}

		locator.Register("test_middleware", testMiddleware)

		mw, err := locator.Get("test_middleware")
		require.NoError(t, err)
		assert.Equal(t, testMiddleware, mw)
	})

	t.Run("get non-existing middleware", func(t *testing.T) {
		locator := middleware.NewMiddlewareLocator()

		mw, err := locator.Get("non_existing")
		require.Error(t, err)
		assert.Nil(t, mw)
		assert.Contains(t, err.Error(), "no middleware with name non_existing found")
	})
}
