package amqp

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	dsn         string
	conn        *amqp.Connection
	lock        sync.Mutex
	channelPool *ChannelPool
}

func NewConnection(dsn string) (*Connection, error) {
	conn := &Connection{dsn: dsn}
	err := conn.connect()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Connection) Channel() (*amqp.Channel, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn == nil || c.conn.IsClosed() {
		if err := c.connect(); err != nil {
			return nil, err
		}
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return ch, nil
}

func (c *Connection) GetChannel() (*amqp.Channel, error) {
	if c.channelPool != nil {
		return c.channelPool.Get()
	}
	return c.Channel()
}

func (c *Connection) PutChannel(ch *amqp.Channel) {
	if c.channelPool != nil {
		c.channelPool.Put(ch)
	} else if ch != nil {
		_ = ch.Close()
	}
}

func (c *Connection) InitChannelPool(size int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.channelPool != nil {
		_ = c.channelPool.Close()
	}

	c.channelPool = NewChannelPool(c, size)
}

func (c *Connection) CloseChannelPool() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.channelPool != nil {
		err := c.channelPool.Close()
		c.channelPool = nil
		return err
	}
	return nil
}

func (c *Connection) connect() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.connectInternal()
}

func (c *Connection) connectInternal() error {
	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP broker at '%s': %w", c.dsn, err)
	}

	c.conn = conn

	return nil
}
