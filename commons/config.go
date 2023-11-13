package commons

import (
	"time"
)

type DiscMode string

var (
	ModeBootstrapper DiscMode = "boot"
	ModeServer       DiscMode = "server"
	ModeClient       DiscMode = "client"
)

type Config struct {
	PrivateKey string `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`

	ListenAddrs []string `json:"listenAddrs" yaml:"listenAddrs"`

	PSK         []byte        `json:"psk,omitempty" yaml:"psk,omitempty"`
	DialTimeout time.Duration `json:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty"`
	DisablePing bool          `json:"disablePing,omitempty" yaml:"disablePing,omitempty"`
	UserAgent   string        `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`

	NatPortMap bool `json:"natPortMap,omitempty" yaml:"natPortMap,omitempty"`
	AutoNat    bool `json:"autoNat,omitempty" yaml:"autoNat,omitempty"`

	ConnManager *ConnManagerConfig `json:"connManager,omitempty" yaml:"connManager,omitempty"`
	Discovery   *DiscoveryConfig   `json:"discovery" yaml:"discovery"`
	Pubsub      *PubsubConfig      `json:"pubsub" yaml:"pubsub"`

	MdnsTag string `json:"mdnsTag,omitempty" yaml:"mdnsTag,omitempty"`

	MsgRouterQueueSize int `json:"msgRouterQueueSize,omitempty" yaml:"msgRouterQueueSize,omitempty"`
	MsgRouterWorkers   int `json:"msgRouterWorkers,omitempty" yaml:"msgRouterWorkers,omitempty"`

	PProf *PProf `json:"pprof,omitempty" yaml:"pprof,omitempty"`
}

func (c *Config) Defaults() {
	if c.DialTimeout.Milliseconds() == 0 {
		c.DialTimeout = time.Second * 15
	}
	if c.MsgRouterQueueSize == 0 {
		c.MsgRouterQueueSize = 1024
	}
	if c.MsgRouterWorkers == 0 {
		c.MsgRouterWorkers = 10
	}
}

type ConnManagerConfig struct {
	LowWaterMark  int           `json:"lowWaterMark" yaml:"lowWaterMark"`
	HighWaterMark int           `json:"highWaterMark" yaml:"highWaterMark"`
	GracePeriod   time.Duration `json:"gracePeriod" yaml:"gracePeriod"`
}

func (cm *ConnManagerConfig) Defaults() {
	if cm.LowWaterMark == 0 {
		cm.LowWaterMark = 5
	}
	if cm.HighWaterMark == 0 {
		cm.HighWaterMark = 25
	}
	if cm.GracePeriod.Milliseconds() == 0 {
		cm.GracePeriod = time.Minute
	}
}

type DiscoveryConfig struct {
	Mode           DiscMode `json:"mode" yaml:"mode"`
	Bootstrappers  []string `json:"bootstrappers" yaml:"bootstrappers"`
	ProtocolPrefix string   `json:"protocolPrefix,omitempty" yaml:"protocolPrefix,omitempty"`
}

func (dc *DiscoveryConfig) Defaults() {
	if len(dc.Mode) == 0 {
		dc.Mode = ModeServer
	}
	if len(dc.ProtocolPrefix) == 0 {
		dc.ProtocolPrefix = "p2pmq"
	}
}

type PProf struct {
	Enabled bool
	Port    uint
}
