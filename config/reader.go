package config

import (
	"os"
)

type ConfigReader interface {
	Read(path string, processors ...ConfigProcessor) ([]byte, error)
}

type FileConfigReader struct{}

func (r *FileConfigReader) Read(path string, processors ...ConfigProcessor) ([]byte, error) {
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
