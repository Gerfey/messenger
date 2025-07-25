package amqp

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	dsn  string
	conn *amqp.Connection
	lock sync.Mutex
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

func (c *Connection) connect() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to AMQP broker at '%s': %w", c.dsn, err)
	}

	c.conn = conn

	return nil
}
