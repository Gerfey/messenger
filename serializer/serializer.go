package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/envelope"
)

type Serializer struct {
	resolver api.TypeResolver
}

func NewSerializer(resolver api.TypeResolver) api.Serializer {
	return &Serializer{resolver: resolver}
}

func (s *Serializer) Marshal(env api.Envelope) ([]byte, map[string]string, error) {
	msg := env.Message()
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal body: %w", err)
	}

	headers := map[string]string{
		"type": reflect.TypeOf(msg).String(),
	}

	stamps := env.Stamps()
	if len(stamps) > 0 {
		var serializedStamps []config.SerializedStamp
		for _, stamp := range stamps {
			data, err := json.Marshal(stamp)
			if err != nil {
				return nil, nil, fmt.Errorf("marshal stamp: %w", err)
			}

			serializedStamps = append(serializedStamps, config.SerializedStamp{
				Type: reflect.TypeOf(stamp).String(),
				Data: data,
			})
		}

		stampsJSON, err := json.Marshal(serializedStamps)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal stamps: %w", err)
		}

		headers["stamps"] = string(stampsJSON)
	}

	return body, headers, nil
}

func (s *Serializer) Unmarshal(body []byte, headers map[string]string) (api.Envelope, error) {
	typeName, ok := headers["type"]
	if !ok {
		return nil, fmt.Errorf("missing 'type' header")
	}

	msgType, err := s.resolver.ResolveMessageType(typeName)
	if err != nil {
		return nil, err
	}

	msgPtr := reflect.New(msgType.Elem()).Interface()
	if err := json.Unmarshal(body, msgPtr); err != nil {
		return nil, fmt.Errorf("unmarshal message body: %w", err)
	}

	env := envelope.NewEnvelope(msgPtr)

	if rawStamps, ok := headers["stamps"]; ok {
		var sStamps []config.SerializedStamp
		if err := json.Unmarshal([]byte(rawStamps), &sStamps); err != nil {
			return nil, fmt.Errorf("unmarshal stamps array: %w", err)
		}

		for _, sStamp := range sStamps {
			t, err := s.resolver.ResolveStampType(sStamp.Type)
			if err != nil {
				continue
			}

			var stampValue any
			if t.Kind() == reflect.Ptr {
				stampValue = reflect.New(t.Elem()).Interface()
				if err := json.Unmarshal(sStamp.Data, stampValue); err != nil {
					continue
				}
			} else {
				ptrValue := reflect.New(t)
				if err := json.Unmarshal(sStamp.Data, ptrValue.Interface()); err != nil {
					continue
				}
				stampValue = ptrValue.Elem().Interface()
			}

			env = env.WithStamp(stampValue.(api.Stamp))
		}
	}

	return env, nil
}
