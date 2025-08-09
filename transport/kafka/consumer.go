package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/stamps"
)

const (
	minBytes          = 10e3 // 10KB
	maxBytes          = 10e6 // 10MB
	sessionTimeout    = 10 * time.Second
	rebalanceTimeout  = 5 * time.Second
	heartbeatInterval = 2 * time.Second
	defaultPoolSize   = 10
	readLagInterval   = -1

	workerPoolCheckInterval = 30 * time.Second
	workerBatchSize         = 5
	errorBackoffDelay       = 100 * time.Millisecond
)

type Consumer struct {
	cfg             TransportConfig
	serializer      api.Serializer
	conn            *Connection
	readers         []*kafka.Reader
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	batchMutex      sync.Mutex
	batchMessages   []kafka.Message
	deferredCommits sync.Map
	logger          *slog.Logger
}

type messageWithReader struct {
	message kafka.Message
	reader  *kafka.Reader
}

func NewConsumer(cfg TransportConfig, ser api.Serializer, conn *Connection, logger *slog.Logger) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		cfg:        cfg,
		serializer: ser,
		conn:       conn,
		readers:    make([]*kafka.Reader, 0),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, api.Envelope) error) error {
	for _, topic := range c.cfg.Options.Topics {
		readerConfig := kafka.ReaderConfig{
			GroupID:           c.cfg.Options.Group,
			Topic:             topic,
			CommitInterval:    c.cfg.Options.Consumer.Commit.Interval,
			MinBytes:          minBytes,
			MaxBytes:          maxBytes,
			ReadLagInterval:   readLagInterval,
			SessionTimeout:    c.cfg.Options.Consumer.SessionTimeout,
			RebalanceTimeout:  rebalanceTimeout,
			HeartbeatInterval: c.cfg.Options.Consumer.HeartbeatInterval,
			MaxWait:           time.Second,
		}

		switch c.cfg.Options.Consumer.Rebalance.Strategy {
		case "range":
			readerConfig.GroupBalancers = []kafka.GroupBalancer{kafka.RangeGroupBalancer{}}
		case "roundrobin":
			readerConfig.GroupBalancers = []kafka.GroupBalancer{kafka.RoundRobinGroupBalancer{}}
		}

		c.configureOffset(&readerConfig)
		reader := c.conn.CreateReader(readerConfig)
		c.readers = append(c.readers, reader)
	}

	jobs := make(chan job)
	c.startWorkerPool(ctx, jobs, handler)

	for _, reader := range c.readers {
		c.wg.Add(1)
		go func(r *kafka.Reader) {
			defer c.wg.Done()
			c.fetchMessages(ctx, r, jobs)
		}(reader)
	}

	if c.cfg.Options.Consumer.Commit.Strategy == "batch" {
		go c.startBatchCommitter(ctx)
	}

	<-ctx.Done()

	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.ErrorContext(ctx, "Failed to close Kafka reader", "error", err)
		}
	}

	c.cancel()
	c.wg.Wait()

	return ctx.Err()
}

func (c *Consumer) configureOffset(config *kafka.ReaderConfig) {
	switch c.cfg.Options.Consumer.OffsetConfig.Type {
	case "earliest":
		config.StartOffset = kafka.FirstOffset
	case "specific":
		config.StartOffset = c.cfg.Options.Consumer.OffsetConfig.Value
	default:
		config.StartOffset = kafka.LastOffset
	}
}

func (c *Consumer) startWorkerPool(
	ctx context.Context,
	jobs chan job,
	handler func(context.Context, api.Envelope) error,
) {
	poolSize := c.cfg.Options.Consumer.Pool.Size
	if poolSize <= 0 {
		poolSize = defaultPoolSize
	}

	for range poolSize {
		c.wg.Add(1)
		go c.startWorker(ctx, jobs, handler)
	}
}

func (c *Consumer) startWorker(ctx context.Context, jobs chan job, handler func(context.Context, api.Envelope) error) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-jobs:
			if !ok {
				return
			}
			c.handleMessage(ctx, j.r, j.msg, handler)
		}
	}
}

func (c *Consumer) fetchMessages(ctx context.Context, r *kafka.Reader, jobs chan job) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				c.logger.ErrorContext(ctx, "Failed to fetch message", "error", err)
				time.Sleep(errorBackoffDelay)

				continue
			}

			select {
			case <-ctx.Done():
				return
			case jobs <- job{r: r, msg: msg}:
			}
		}
	}
}

func (c *Consumer) handleMessage(
	ctx context.Context,
	r *kafka.Reader,
	msg kafka.Message,
	handler func(context.Context, api.Envelope) error,
) {
	env, err := c.serializer.Unmarshal(msg.Value, c.headerMap(msg.Headers))
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to unmarshal message", "error", err)

		c.commitMessage(ctx, r, msg)

		return
	}

	env = env.WithStamp(stamps.ReceivedStamp{Transport: c.cfg.Name})

	if handlerErr := handler(ctx, env); handlerErr != nil {
		c.logger.ErrorContext(ctx, "Handler failed", "error", handlerErr)
		c.commitMessage(ctx, r, msg)

		return
	}

	c.commitMessage(ctx, r, msg)
}

func (c *Consumer) commitMessage(ctx context.Context, r *kafka.Reader, msg kafka.Message) {
	switch c.cfg.Options.Consumer.Commit.Strategy {
	case "auto":
		if err := r.CommitMessages(ctx, msg); err != nil {
			c.logger.ErrorContext(ctx, "Failed to commit message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"error", err)
		}
	case "manual":
	case "batch":
		c.batchMutex.Lock()
		c.batchMessages = append(c.batchMessages, msg)

		msgKey := fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
		c.deferredCommits.Store(msgKey, messageWithReader{message: msg, reader: r})

		if len(c.batchMessages) >= c.cfg.Options.Consumer.Commit.BatchSize {
			c.batchMutex.Unlock()
			c.commitBatch(ctx)
		} else {
			c.batchMutex.Unlock()
		}
	case "deferred":
		msgKey := fmt.Sprintf("%s-%d-%d", msg.Topic, msg.Partition, msg.Offset)
		c.deferredCommits.Store(msgKey, messageWithReader{message: msg, reader: r})
	}
}

func (c *Consumer) startBatchCommitter(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.Options.Consumer.Commit.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.commitBatch(ctx)
		}
	}
}

func (c *Consumer) commitBatch(ctx context.Context) {
	c.batchMutex.Lock()
	defer c.batchMutex.Unlock()

	readerMessages := c.groupMessagesByReader()

	for reader, messages := range readerMessages {
		if len(messages) == 0 {
			continue
		}

		partitionOffsets := c.findMaxOffsetPerPartition(messages)

		c.commitMessagesAndCleanup(ctx, reader, messages, partitionOffsets)
	}

	c.batchMessages = c.batchMessages[:0]
}

func (c *Consumer) groupMessagesByReader() map[*kafka.Reader][]kafka.Message {
	readerMessages := make(map[*kafka.Reader][]kafka.Message)

	c.deferredCommits.Range(func(_, value any) bool {
		if msgWithReader, ok := value.(messageWithReader); ok {
			readerMessages[msgWithReader.reader] = append(readerMessages[msgWithReader.reader], msgWithReader.message)
		}

		return true
	})

	return readerMessages
}

func (c *Consumer) findMaxOffsetPerPartition(messages []kafka.Message) map[int]kafka.Message {
	partitionOffsets := make(map[int]kafka.Message)

	for _, msg := range messages {
		currentMax, exists := partitionOffsets[msg.Partition]
		if !exists || currentMax.Offset < msg.Offset {
			partitionOffsets[msg.Partition] = msg
		}
	}

	return partitionOffsets
}

func (c *Consumer) commitMessagesAndCleanup(
	ctx context.Context,
	reader *kafka.Reader,
	messages []kafka.Message,
	partitionOffsets map[int]kafka.Message,
) {
	for _, msg := range partitionOffsets {
		if err := reader.CommitMessages(ctx, msg); err != nil {
			c.logger.ErrorContext(ctx, "Failed to commit message batch",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"error", err.Error())
		} else {
			c.cleanupCommittedMessages(messages, msg)
		}
	}
}

func (c *Consumer) cleanupCommittedMessages(messages []kafka.Message, committedMsg kafka.Message) {
	for _, commitedMsg := range messages {
		if commitedMsg.Partition == committedMsg.Partition && commitedMsg.Offset <= committedMsg.Offset {
			c.deferredCommits.Delete(
				fmt.Sprintf("%s-%d-%d", commitedMsg.Topic, commitedMsg.Partition, commitedMsg.Offset),
			)
		}
	}
}

type job struct {
	r   *kafka.Reader
	msg kafka.Message
}

func (c *Consumer) headerMap(headers []kafka.Header) map[string]string {
	m := make(map[string]string)
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}

	return m
}
