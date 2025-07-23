package transport_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/transport"

	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewLocator(t *testing.T) {
	t.Run("create new locator", func(t *testing.T) {
		locator := transport.NewLocator()

		require.NotNil(t, locator)
		assert.IsType(t, &transport.Locator{}, locator)

		all := locator.GetAllTransports()
		assert.Empty(t, all)
	})
}

func TestLocator_Register(t *testing.T) {
	t.Run("register single transport", func(t *testing.T) {
		locator := transport.NewLocator()
		tr := &helpers.TestTransport{}

		err := locator.Register("test-transport", tr)

		require.NoError(t, err)

		retrieved := locator.GetTransport("test-transport")
		assert.Equal(t, tr, retrieved)
	})

	t.Run("register multiple transports", func(t *testing.T) {
		locator := transport.NewLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		err1 := locator.Register("transport1", transport1)
		err2 := locator.Register("transport2", transport2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		retrieved1 := locator.GetTransport("transport1")
		retrieved2 := locator.GetTransport("transport2")
		assert.Equal(t, transport1, retrieved1)
		assert.Equal(t, transport2, retrieved2)

		all := locator.GetAllTransports()
		assert.Len(t, all, 2)
	})

	t.Run("register transport with same name overwrites previous", func(t *testing.T) {
		locator := transport.NewLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		err1 := locator.Register("test-transport", transport1)
		err2 := locator.Register("test-transport", transport2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		retrieved := locator.GetTransport("test-transport")
		assert.Same(t, transport2, retrieved)
		assert.NotSame(t, transport1, retrieved)

		all := locator.GetAllTransports()
		assert.Len(t, all, 1)
	})

	t.Run("register transport with empty name", func(t *testing.T) {
		locator := transport.NewLocator()
		tr := &helpers.TestTransport{}

		err := locator.Register("", tr)

		require.NoError(t, err)

		retrieved := locator.GetTransport("")
		assert.Equal(t, tr, retrieved)
	})

	t.Run("register nil transport", func(t *testing.T) {
		locator := transport.NewLocator()

		err := locator.Register("test-transport", nil)

		require.NoError(t, err)

		retrieved := locator.GetTransport("test-transport")
		assert.Nil(t, retrieved)
	})
}

func TestLocator_GetTransport(t *testing.T) {
	t.Run("get existing transport", func(t *testing.T) {
		locator := transport.NewLocator()
		tr := &helpers.TestTransport{}

		err := locator.Register("test-transport", tr)
		require.NoError(t, err)

		retrieved := locator.GetTransport("test-transport")
		assert.Equal(t, tr, retrieved)
	})

	t.Run("get non-existing transport", func(t *testing.T) {
		locator := transport.NewLocator()

		retrieved := locator.GetTransport("non-existing")
		assert.Nil(t, retrieved)
	})

	t.Run("get transport with empty name", func(t *testing.T) {
		locator := transport.NewLocator()

		retrieved := locator.GetTransport("")
		assert.Nil(t, retrieved)
	})

	t.Run("get after multiple registrations", func(t *testing.T) {
		locator := transport.NewLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}
		transport3 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("transport1", transport1))
		require.NoError(t, locator.Register("transport2", transport2))
		require.NoError(t, locator.Register("transport3", transport3))

		retrieved1 := locator.GetTransport("transport1")
		retrieved2 := locator.GetTransport("transport2")
		retrieved3 := locator.GetTransport("transport3")

		assert.Equal(t, transport1, retrieved1)
		assert.Equal(t, transport2, retrieved2)
		assert.Equal(t, transport3, retrieved3)

		retrievedNone := locator.GetTransport("non-existing")
		assert.Nil(t, retrievedNone)
	})
}

func TestLocator_GetAllTransports(t *testing.T) {
	t.Run("get all from empty locator", func(t *testing.T) {
		locator := transport.NewLocator()

		all := locator.GetAllTransports()
		assert.Empty(t, all)
	})

	t.Run("get all with single transport", func(t *testing.T) {
		locator := transport.NewLocator()
		tr := &helpers.TestTransport{}

		require.NoError(t, locator.Register("test-transport", tr))

		all := locator.GetAllTransports()
		assert.Len(t, all, 1)
		assert.Contains(t, all, tr)
	})

	t.Run("get all with multiple transports", func(t *testing.T) {
		locator := transport.NewLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}
		transport3 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("transport1", transport1))
		require.NoError(t, locator.Register("transport2", transport2))
		require.NoError(t, locator.Register("transport3", transport3))

		all := locator.GetAllTransports()
		assert.Len(t, all, 3)
		assert.Contains(t, all, transport1)
		assert.Contains(t, all, transport2)
		assert.Contains(t, all, transport3)
	})

	t.Run("get all with nil transport", func(t *testing.T) {
		locator := transport.NewLocator()
		tr := &helpers.TestTransport{}

		require.NoError(t, locator.Register("real-transport", tr))
		require.NoError(t, locator.Register("nil-transport", nil))

		all := locator.GetAllTransports()
		assert.Len(t, all, 2)
		assert.Contains(t, all, tr)
		assert.Contains(t, all, nil)
	})

	t.Run("get all returns slice of transports", func(t *testing.T) {
		locator := transport.NewLocator()
		transport1 := &helpers.TestTransport{}
		transport2 := &helpers.TestTransport{}

		require.NoError(t, locator.Register("transport1", transport1))
		require.NoError(t, locator.Register("transport2", transport2))

		all1 := locator.GetAllTransports()
		all2 := locator.GetAllTransports()

		assert.ElementsMatch(t, all1, all2)

		assert.NotSame(t, &all1, &all2)
	})
}

func TestLocator_Integration(t *testing.T) {
	t.Run("full workflow with multiple operations", func(t *testing.T) {
		locator := transport.NewLocator()

		all := locator.GetAllTransports()
		assert.Empty(t, all)

		defaultTransport := &helpers.TestTransport{}
		asyncTransport := &helpers.TestTransport{}

		require.NoError(t, locator.Register("default", defaultTransport))
		require.NoError(t, locator.Register("async", asyncTransport))

		all = locator.GetAllTransports()
		assert.Len(t, all, 2)

		retrievedDefault := locator.GetTransport("default")
		retrievedAsync := locator.GetTransport("async")
		assert.Equal(t, defaultTransport, retrievedDefault)
		assert.Equal(t, asyncTransport, retrievedAsync)

		newDefaultTransport := &helpers.TestTransport{}
		require.NoError(t, locator.Register("default", newDefaultTransport))

		all = locator.GetAllTransports()
		assert.Len(t, all, 2)

		retrievedNewDefault := locator.GetTransport("default")
		assert.Same(t, newDefaultTransport, retrievedNewDefault)
		assert.NotSame(t, defaultTransport, retrievedNewDefault)

		retrievedAsync = locator.GetTransport("async")
		assert.Same(t, asyncTransport, retrievedAsync)
	})
}
