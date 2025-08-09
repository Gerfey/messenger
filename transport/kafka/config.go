package kafka

import "time"

type TransportConfig struct {
	Name    string
	DSN     string
	Options OptionsConfig
}

type OptionsConfig struct {
	Topics   []string              `yaml:"topics,omitempty"`
	Group    string                `yaml:"group"            default:"default-group"`
	Producer ProducerOptionsConfig `yaml:"producer"`
	Consumer ConsumerOptionsConfig `yaml:"consumer"`
	Key      KeyConfig             `yaml:"key"`
}

type ProducerOptionsConfig struct {
	Async             bool          `yaml:"async"               default:"false"`
	AutoTopicCreation bool          `yaml:"auto_topic_creation" default:"false"`
	RequiredAcks      int           `yaml:"required_acks"       default:"1"` // 0, 1, -1 (all)
	BatchSize         int           `yaml:"batch_size"          default:"256"`
	BatchTimeout      time.Duration `yaml:"batch_timeout"       default:"5ms"`
	WriteTimeout      time.Duration `yaml:"write_timeout"       default:"10s"`
	ReadTimeout       time.Duration `yaml:"read_timeout"        default:"10s"`
	Balancer          string        `yaml:"balancer"            default:"round_robin"` // least_bytes, hash, round_robin
}

type ConsumerOptionsConfig struct {
	OffsetConfig      OffsetConfig    `yaml:"offset"`
	Commit            CommitConfig    `yaml:"commit"`
	Pool              PoolConfig      `yaml:"pool"`
	Rebalance         RebalanceConfig `yaml:"rebalance"`
	SessionTimeout    time.Duration   `yaml:"session_timeout"    default:"10s"`
	HeartbeatInterval time.Duration   `yaml:"heartbeat_interval" default:"2s"`
}

type OffsetConfig struct {
	Type  string `yaml:"type"  default:"latest"` // earliest, latest, specific
	Value int64  `yaml:"value"`
}

type CommitConfig struct {
	Strategy  string        `yaml:"strategy"   default:"batch"` // auto, manual, batch, deferred
	Interval  time.Duration `yaml:"interval"   default:"500ms"` // only for batch
	BatchSize int           `yaml:"batch_size" default:"10"`    // only for batch
}

type PoolConfig struct {
	Size    int `yaml:"size"     default:"3"`
	MinSize int `yaml:"min_size" default:"2"`
	MaxSize int `yaml:"max_size" default:"10"`
}

type RebalanceConfig struct {
	Strategy string `yaml:"strategy" default:"range"` // range, roundrobin
}

type KeyConfig struct {
	Strategy string `yaml:"strategy" default:"none"` // none, message_id
}
