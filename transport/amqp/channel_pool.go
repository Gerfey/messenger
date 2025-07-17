package amqp

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelPool struct {
	conn     *Connection
	channels chan *amqp.Channel
	size     int
	mu       sync.RWMutex
	closed   bool
}

func NewChannelPool(conn *Connection, size int) *ChannelPool {
	if size <= 0 {
		size = 10
	}

	pool := &ChannelPool{
		conn:     conn,
		channels: make(chan *amqp.Channel, size),
		size:     size,
	}

	return pool
}

func (p *ChannelPool) Get() (*amqp.Channel, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("channel pool is closed")
	}
	p.mu.RUnlock()

	select {
	case ch := <-p.channels:
		if ch != nil && !ch.IsClosed() {
			return ch, nil
		}
		ch, err := p.createNewChannel()
		if err != nil {
			return nil, fmt.Errorf("failed to create new channel: %w", err)
		}
		return ch, nil
	default:
		ch, err := p.createNewChannel()
		if err != nil {
			return nil, fmt.Errorf("failed to create new channel: %w", err)
		}
		return ch, nil
	}
}

func (p *ChannelPool) createNewChannel() (*amqp.Channel, error) {
	p.conn.lock.Lock()
	defer p.conn.lock.Unlock()

	if p.conn.conn == nil || p.conn.conn.IsClosed() {
		if err := p.conn.connectInternal(); err != nil {
			return nil, err
		}
	}

	ch, err := p.conn.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return ch, nil
}

func (p *ChannelPool) Put(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		_ = ch.Close()
		return
	}
	p.mu.RUnlock()

	select {
	case p.channels <- ch:
	default:
		_ = ch.Close()
	}
}

func (p *ChannelPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.channels)

	for ch := range p.channels {
		if ch != nil && !ch.IsClosed() {
			_ = ch.Close()
		}
	}

	return nil
}
