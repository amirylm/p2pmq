package commons

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
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

	PProf *PProf `json:"pprof,omitempty" yaml:"pprof,omitempty"`
}

func (c *Config) Defaults() {
	if c.DialTimeout.Milliseconds() == 0 {
		c.DialTimeout = time.Second * 15
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

// PubsubConfig contains config for the pubsub router
type PubsubConfig struct {
	Topics         []TopicConfig       `json:"topics" yaml:"topics"`
	Overlay        *OverlayParams      `json:"overlay,omitempty" yaml:"overlay,omitempty"`
	SubFilter      *SubscriptionFilter `json:"subFilter,omitempty" yaml:"subFilter,omitempty"`
	MaxMessageSize int                 `json:"maxMessageSize,omitempty" yaml:"maxMessageSize,omitempty"`
	Trace          bool                `json:"trace,omitempty" yaml:"trace,omitempty"`
}

// TopicConfig contains configuration of a pubsub topic
type TopicConfig struct {
	Name         string         `json:"name" yaml:"name"`
	BufferSize   int            `json:"bufferSize,omitempty" yaml:"bufferSize,omitempty"`
	RateLimit    float64        `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	Overlay      *OverlayParams `json:"overlay,omitempty" yaml:"overlay,omitempty"`
	MsgValidator string         `json:"msgValidator,omitempty" yaml:"msgValidator,omitempty"`
	Scoring      *ScoringParams `json:"scoring,omitempty" yaml:"scoring,omitempty"`
}

// SubscriptionFilter configurations
type SubscriptionFilter struct {
	// Pattern of topics to accept
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	// Limit is the max number of topics to accept
	Limit int `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// OverlayParams are the overlay params as defined in
// https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md#parameters
type OverlayParams struct {
	// D is the desired outbound degree of the network (6)
	D int32 `json:"d,omitempty" yaml:"d,omitempty"`
	// Dlow is the lower bound for outbound degree	(4)
	Dlow int32 `json:"dlo,omitempty" yaml:"dlo,omitempty"`
	// Dhi is the upper bound for outbound degree (12)
	Dhi int32 `json:"dhi,omitempty" yaml:"dhi,omitempty"`
	// Dlazy is the outbound degree for gossip emission (D)
	Dlazy int32 `json:"dlazy,omitempty" yaml:"dlazy,omitempty"`
	// HeartbeatInterval is the time between heartbeats (1sec)
	HeartbeatInterval time.Duration `json:"heartbeatInterval,omitempty" yaml:"heartbeatInterval,omitempty"`
	// FanoutTtl time-to-live for each topic's fanout state (60sec)
	FanoutTtl time.Duration `json:"fanoutTTL,omitempty" yaml:"fanoutTTL,omitempty"`
	// McacheLen is the number of history windows in message cache (5)
	McacheLen int32 `json:"mCacheLen,omitempty" yaml:"mCacheLen,omitempty"`
	// McacheGossip is the number of history windows to use when emitting gossip (3)
	McacheGossip int32 `json:"mCacheGossip,omitempty" yaml:"mCacheGossip,omitempty"`
	// SeenTtl is the expiry time for cache of seen message ids (2min)
	SeenTtl time.Duration `json:"seenTTL,omitempty" yaml:"seenTTL,omitempty"`
}

type ScoringParams struct {
	// TODO
}

type PProf struct {
	Enabled bool
	Port    uint
}

func ReadConfig(cfgAddr string) (*Config, error) {
	raw, err := os.ReadFile(cfgAddr)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	cfg := Config{}
	if strings.Contains(cfgAddr, ".yaml") || strings.Contains(cfgAddr, ".yml") {
		err = yaml.Unmarshal(raw, &cfg)
	} else {
		err = json.Unmarshal(raw, &cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}
	return &cfg, nil
}
