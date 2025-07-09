package config

type ConfigProcessor interface {
	Process(content []byte) ([]byte, error)
}
