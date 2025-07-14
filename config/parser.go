package config

import (
	"bytes"

	"github.com/creasty/defaults"
	"github.com/ilyakaznacheev/cleanenv"
)

type ConfigParser interface {
	Parse(content []byte, cfg interface{}) error
}

type YAMLParser struct{}

func (p *YAMLParser) Parse(content []byte, cfg interface{}) error {
	reader := bytes.NewReader(content)

	if errParseYAML := cleanenv.ParseYAML(reader, cfg); errParseYAML != nil {
		return errParseYAML
	}

	if errDefault := defaults.Set(cfg); errDefault != nil {
		return errDefault
	}

	return nil
}
