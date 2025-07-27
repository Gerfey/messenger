package transport

import (
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/core/stamps"
)

type SenderLocator struct {
	senders    map[string]api.Sender
	sendersMap map[reflect.Type][]string
	fallback   []string
}

func NewSenderLocator() api.SenderLocator {
	return &SenderLocator{
		senders:    make(map[string]api.Sender),
		sendersMap: make(map[reflect.Type][]string),
		fallback:   make([]string, 0),
	}
}

func (r *SenderLocator) Register(name string, sender api.Sender) error {
	r.senders[name] = sender

	return nil
}

func (r *SenderLocator) RegisterMessageType(messageType reflect.Type, senderAliases []string) {
	r.sendersMap[messageType] = senderAliases
}

func (r *SenderLocator) SetFallback(senderAliases []string) {
	r.fallback = senderAliases
}

func (r *SenderLocator) GetSenders(env api.Envelope) []api.Sender {
	var senders []api.Sender
	seen := make(map[string]bool)

	if senders = r.getSendersFromStamp(env, seen); len(senders) > 0 {
		return senders
	}

	messageType := reflect.TypeOf(env.Message())
	found := r.getSendersFromExactType(messageType, &senders, seen)

	if !found {
		found = r.getSendersFromAssignableTypes(messageType, &senders, seen)
	}

	if !found {
		r.getSendersFromFallback(&senders, seen)
	}

	return senders
}

func (r *SenderLocator) getSendersFromStamp(env api.Envelope, seen map[string]bool) []api.Sender {
	var senders []api.Sender

	if transportsStamp, ok := envelope.LastStampOf[stamps.TransportNameStamp](env); ok {
		for _, transportName := range transportsStamp.Transports {
			if sender, exists := r.senders[transportName]; exists && !seen[transportName] {
				senders = append(senders, sender)
				seen[transportName] = true
			}
		}
	}

	return senders
}

func (r *SenderLocator) getSendersFromExactType(
	messageType reflect.Type,
	senders *[]api.Sender,
	seen map[string]bool,
) bool {
	found := false

	if senderAliases, exists := r.sendersMap[messageType]; exists {
		for _, alias := range senderAliases {
			if sender, senderExists := r.senders[alias]; senderExists && !seen[alias] {
				*senders = append(*senders, sender)
				seen[alias] = true
				found = true
			}
		}
	}

	return found
}

func (r *SenderLocator) getSendersFromAssignableTypes(
	messageType reflect.Type,
	senders *[]api.Sender,
	seen map[string]bool,
) bool {
	found := false

	for msgType, senderAliases := range r.sendersMap {
		if messageType.AssignableTo(msgType) {
			for _, alias := range senderAliases {
				if sender, senderExists := r.senders[alias]; senderExists && !seen[alias] {
					*senders = append(*senders, sender)
					seen[alias] = true
					found = true
				}
			}
		}
	}

	return found
}

func (r *SenderLocator) getSendersFromFallback(senders *[]api.Sender, seen map[string]bool) {
	for _, alias := range r.fallback {
		if sender, exists := r.senders[alias]; exists && !seen[alias] {
			*senders = append(*senders, sender)
			seen[alias] = true
		}
	}
}
