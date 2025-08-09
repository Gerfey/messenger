package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	connectionTimeout = 5 * time.Second
)

type Connection struct {
	brokers []string
	dialer  *kafka.Dialer
}

func NewConnection(brokers []string) (*Connection, error) {
	conn := &Connection{
		brokers: brokers,
		dialer: &kafka.Dialer{
			Timeout:   connectionTimeout,
			DualStack: true,
		},
	}

	if err := conn.Check(connectionTimeout); err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Connection) Check(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, broker := range c.brokers {
		conn, connErr := c.dialer.DialContext(ctx, "tcp", broker)
		if connErr != nil {
			return fmt.Errorf("failed to connect to Kafka broker at '%s': %w", broker, connErr)
		}

		if closeErr := conn.Close(); closeErr != nil {
			return fmt.Errorf("failed to close connection to Kafka broker at '%s': %w", broker, closeErr)
		}
	}

	return nil
}

func (c *Connection) CreateReader(config kafka.ReaderConfig) *kafka.Reader {
	config.Brokers = c.brokers
	config.Dialer = c.dialer

	return kafka.NewReader(config)
}

func (c *Connection) CreateWriter(
	topic string,
	opts ProducerOptionsConfig,
	async bool,
	balancer kafka.Balancer,
) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(c.brokers...),
		Topic:                  topic,
		RequiredAcks:           kafka.RequiredAcks(opts.RequiredAcks),
		Async:                  async,
		AllowAutoTopicCreation: opts.AutoTopicCreation,
		Balancer:               balancer,
		BatchSize:              opts.BatchSize,
		BatchTimeout:           opts.BatchTimeout,
		WriteTimeout:           opts.WriteTimeout,
		ReadTimeout:            opts.ReadTimeout,
	}
}
