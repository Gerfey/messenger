package kafka

import "time"

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	Topics       []string        `yaml:"topics,omitempty"`
	Group        string          `yaml:"group"            default:"group"`
	OffsetConfig OffsetConfig    `yaml:"offset_config"`
	Commit       CommitConfig    `yaml:"commit"`
	Pool         PoolConfig      `yaml:"pool"`
	Rebalance    RebalanceConfig `yaml:"rebalance"`
	Key          KeyConfig       `yaml:"key"`
}

type OffsetConfig struct {
	Type  string `yaml:"type"  default:"latest"` // earliest, latest, specific
	Value int64  `yaml:"value"`
}

type CommitConfig struct {
	Strategy  string        `yaml:"strategy"   default:"auto"`
	Interval  time.Duration `yaml:"interval"   default:"1s"`
	BatchSize int           `yaml:"batch_size" default:"100"`
}

type PoolConfig struct {
	Size    int  `yaml:"size"     default:"10"`
	MinSize int  `yaml:"min_size" default:"5"`
	MaxSize int  `yaml:"max_size" default:"50"`
	Dynamic bool `yaml:"dynamic"  default:"false"`
}

type RebalanceConfig struct {
	Strategy string `yaml:"strategy" default:"range"` // range, roundrobin
}

type KeyConfig struct {
	Strategy string `yaml:"strategy" default:"none"` // none, message_id
}
