package helpers

type TestMessage struct {
	ID      string
	Content string
}

type SimpleMessage string

type ComplexMessage struct {
	ID       string
	Type     string
	Metadata map[string]string
	Payload  any
}

type TestStamp struct {
	Value string
}

type AnotherStamp struct {
	Number int
}

func NewTestMessage(id, content string) *TestMessage {
	return &TestMessage{
		ID:      id,
		Content: content,
	}
}

func NewComplexMessage(id, msgType string) *ComplexMessage {
	return &ComplexMessage{
		ID:       id,
		Type:     msgType,
		Metadata: make(map[string]string),
		Payload:  nil,
	}
}
