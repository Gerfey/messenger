package kafka

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gerfey/messenger/config"
)

type TransportConfig struct {
	Name             string
	Brokers          []string
	Topic            string
	GroupID          string
	Offset           string
	ConsumerPoolSize int
	CommitInterval   time.Duration
}

func NewConfig(name, dsn string, opts config.OptionsConfig) (*TransportConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn: %w", err)
	}

	brokers := strings.Split(u.Host, ",")
	topic := strings.TrimPrefix(u.Path, "/")

	query := u.Query()
	groupID := query.Get("group")
	if groupID == "" {
		return nil, errors.New("groupID is required (e.g. ?group=my-group)")
	}

	offset := query.Get("offset")
	if offset == "" {
		offset = "latest"
	}

	consumerPoolSize := 1
	if opts.ConsumerPoolSize > 0 {
		consumerPoolSize = opts.ConsumerPoolSize
	}

	commitInterval := 1 * time.Second
	if opts.CommitInterval > 0 {
		commitInterval = opts.CommitInterval
	}

	return &TransportConfig{
		Name:             name,
		Brokers:          brokers,
		Topic:            topic,
		GroupID:          groupID,
		Offset:           offset,
		ConsumerPoolSize: consumerPoolSize,
		CommitInterval:   commitInterval,
	}, nil
}
