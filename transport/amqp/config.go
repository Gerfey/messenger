package amqp

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	AutoSetup bool             `yaml:"auto_setup" default:"false"`
	Pool      PoolConfig       `yaml:"pool"`
	Exchange  ExchangeConfig   `yaml:"exchange"`
	Queues    map[string]Queue `yaml:"queues"`
}

type PoolConfig struct {
	Size    int `yaml:"size"     default:"10"`
	MinSize int `yaml:"min_size" default:"5"`
	MaxSize int `yaml:"max_size" default:"20"`
}

type ExchangeConfig struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"        default:"topic"` // topic, direct, fanout
	Durable    bool   `yaml:"durable"     default:"true"`
	AutoDelete bool   `yaml:"auto_delete" default:"false"`
	Internal   bool   `yaml:"internal"    default:"false"`
}

type Queue struct {
	BindingKeys []string `yaml:"binding_keys"`
	Durable     bool     `yaml:"durable"      default:"true"`
	Exclusive   bool     `yaml:"exclusive"    default:"false"`
	AutoDelete  bool     `yaml:"auto_delete"  default:"false"`
}
