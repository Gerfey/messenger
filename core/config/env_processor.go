package config

import (
	"os"
	"regexp"
	"strings"
)

const (
	expectedMatchParts = 2
)

type EnvVarProcessor struct{}

func (p *EnvVarProcessor) Process(content []byte) ([]byte, error) {
	contentStr := string(content)
	re := regexp.MustCompile(`%env\(([A-Z0-9_]+)\)%`)
	matches := re.FindAllStringSubmatch(contentStr, -1)

	for _, match := range matches {
		if len(match) == expectedMatchParts {
			envName := match[1]
			envValue := os.Getenv(envName)
			contentStr = strings.ReplaceAll(contentStr, match[0], envValue)
		}
	}

	return []byte(contentStr), nil
}
