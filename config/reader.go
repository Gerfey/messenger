package config

import (
	"os"
)

type Processor interface {
	Process(content []byte) ([]byte, error)
}

type Reader interface {
	Read(path string, processors ...Processor) ([]byte, error)
}

type FileReader struct{}

func (r *FileReader) Read(path string, processors ...Processor) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	for _, processor := range processors {
		content, err = processor.Process(content)
		if err != nil {
			return nil, err
		}
	}

	return content, nil
}
