package redis

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	AutoSetup bool   `yaml:"auto_setup" default:"true"`
	Stream    string `yaml:"stream"     default:"messages"`
	Group     string `yaml:"group"      default:"default"`
	Consumer  string `yaml:"consumer"   default:"consumer"`
}
