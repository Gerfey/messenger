package redis

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

const connectionTimeout = 5 * time.Second

type Connection struct {
	client *redis.Client
}

func NewConnection(dsn string) (*Connection, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	opts := &redis.Options{
		Addr: u.Host,
	}

	if u.User != nil {
		if password, hasPassword := u.User.Password(); hasPassword {
			opts.Password = password
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	if errPing := client.Ping(ctx).Err(); errPing != nil {
		return nil, fmt.Errorf("ping failed: %w", errPing)
	}

	return &Connection{client: client}, nil
}

func (c *Connection) Client() *redis.Client {
	return c.client
}
