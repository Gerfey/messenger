package serializer

import (
	"encoding/json"
	"errors"
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
		return nil, nil, err
	}

	headers := map[string]string{
		"type": reflect.TypeOf(msg).String(),
	}

	stamps := env.Stamps()
	if len(stamps) > 0 {
		var serializedStamps []config.SerializedStamp
		for _, stamp := range stamps {
			data, marshalErr := json.Marshal(stamp)
			if marshalErr != nil {
				return nil, nil, marshalErr
			}

			serializedStamps = append(serializedStamps, config.SerializedStamp{
				Type: reflect.TypeOf(stamp).String(),
				Data: data,
			})
		}

		stampsJSON, stampsErr := json.Marshal(serializedStamps)
		if stampsErr != nil {
			return nil, nil, stampsErr
		}

		headers["stamps"] = string(stampsJSON)
	}

	return body, headers, nil
}

func (s *Serializer) Unmarshal(body []byte, headers map[string]string) (api.Envelope, error) {
	typeName, ok := headers["type"]
	if !ok {
		return nil, errors.New("missing 'type' header")
	}

	msgType, err := s.resolver.ResolveMessageType(typeName)
	if err != nil {
		return nil, err
	}

	msgPtr := reflect.New(msgType.Elem()).Interface()
	if unmarshalErr := json.Unmarshal(body, msgPtr); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	env := envelope.NewEnvelope(msgPtr)

	if rawStamps, stampsOk := headers["stamps"]; stampsOk {
		env = s.processStamps(env, rawStamps)
	}

	return env, nil
}

func (s *Serializer) processStamps(env api.Envelope, rawStamps string) api.Envelope {
	var sStamps []config.SerializedStamp
	if stampsUnmarshalErr := json.Unmarshal([]byte(rawStamps), &sStamps); stampsUnmarshalErr != nil {
		return env
	}

	for _, sStamp := range sStamps {
		if stamp := s.deserializeStamp(sStamp); stamp != nil {
			env = env.WithStamp(stamp)
		}
	}

	return env
}

func (s *Serializer) deserializeStamp(sStamp config.SerializedStamp) api.Stamp {
	t, resolveErr := s.resolver.ResolveStampType(sStamp.Type)
	if resolveErr != nil {
		return nil
	}

	stampValue := s.createStampValue(t, sStamp.Data)
	if stampValue == nil {
		return nil
	}

	if stamp, stampOk := stampValue.(api.Stamp); stampOk {
		return stamp
	}

	return nil
}

func (s *Serializer) createStampValue(t reflect.Type, data []byte) any {
	var stampValue any

	if t.Kind() == reflect.Ptr {
		stampValue = reflect.New(t.Elem()).Interface()
		if unmarshalErr := json.Unmarshal(data, stampValue); unmarshalErr != nil {
			return nil
		}
	} else {
		ptrValue := reflect.New(t)
		if unmarshalErr := json.Unmarshal(data, ptrValue.Interface()); unmarshalErr != nil {
			return nil
		}
		stampValue = ptrValue.Elem().Interface()
	}

	return stampValue
}
