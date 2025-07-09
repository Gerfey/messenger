package config

import (
	"os"
	"regexp"
	"strings"
)

type EnvVarProcessor struct{}

func (p *EnvVarProcessor) Process(content []byte) ([]byte, error) {
	contentStr := string(content)
	re := regexp.MustCompile(`%env\(([A-Z0-9_]+)\)%`)
	matches := re.FindAllStringSubmatch(contentStr, -1)

	for _, match := range matches {
		if len(match) == 2 {
			envName := match[1]
			envValue := os.Getenv(envName)
			contentStr = strings.Replace(contentStr, match[0], envValue, -1)
		}
	}

	return []byte(contentStr), nil
}
