package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/core/envelope"
)

type TestSerializer struct{}

func NewTestSerializer() api.Serializer {
	return &TestSerializer{}
}

func (s *TestSerializer) Marshal(env api.Envelope) ([]byte, map[string]string, error) {
	msg := env.Message()
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, nil, err
	}

	headers := map[string]string{
		"type": reflect.TypeOf(msg).String(),
	}

	return body, headers, nil
}

func (s *TestSerializer) Unmarshal(_ []byte, _ map[string]string) (api.Envelope, error) {
	fmt.Println("TestJsonSerializer.Unmarshal")

	env := envelope.NewEnvelope("")

	return env, nil
}
