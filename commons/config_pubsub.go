package commons

import (
	"net"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PubsubConfig contains config for the pubsub router
type PubsubConfig struct {
	Topics               []TopicConfig        `json:"topics" yaml:"topics"`
	Overlay              *OverlayParams       `json:"overlay,omitempty" yaml:"overlay,omitempty"`
	SubFilter            *SubscriptionFilter  `json:"subFilter,omitempty" yaml:"subFilter,omitempty"`
	MaxMessageSize       int                  `json:"maxMessageSize,omitempty" yaml:"maxMessageSize,omitempty"`
	MsgValidationTimeout time.Duration        `json:"msgValidationTimeout,omitempty" yaml:"msgValidationTimeout,omitempty"`
	Scoring              *ScoringParams       `json:"scoring,omitempty" yaml:"scoring,omitempty"`
	MsgValidator         *MsgValidationConfig `json:"msgValidator,omitempty" yaml:"msgValidator,omitempty"`
	MsgIDFnConfig        *MsgIDFnConfig       `json:"msgIDFn,omitempty" yaml:"msgIDFn,omitempty"`
	Trace                *PubsubTraceConfig   `json:"trace,omitempty" yaml:"trace,omitempty"`
}

func (psc PubsubConfig) GetTopicConfig(name string) (TopicConfig, bool) {
	for _, t := range psc.Topics {
		if t.Name == name {
			return t, true
		}
	}
	return TopicConfig{}, false
}

type PubsubTraceConfig struct {
	Skiplist []string `json:"skiplist,omitempty" yaml:"skiplist,omitempty"`
	JsonFile string   `json:"jsonFile,omitempty" yaml:"jsonFile,omitempty"`
	Debug    bool     `json:"debug,omitempty" yaml:"debug,omitempty"`
}

type MsgIDFnConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Size int    `json:"size,omitempty" yaml:"size,omitempty"`
}

type MsgValidationConfig struct {
	Timeout     time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Concurrency int           `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
}

func (mvc *MsgValidationConfig) Defaults(other *MsgValidationConfig) *MsgValidationConfig {
	if mvc.Timeout.Milliseconds() == 0 {
		mvc.Timeout = time.Second * 5
	}
	if mvc.Concurrency == 0 {
		mvc.Concurrency = 10
	}
	if other != nil {
		if other.Timeout.Milliseconds() > 0 {
			mvc.Timeout = other.Timeout
		}
		if other.Concurrency > 0 {
			mvc.Concurrency = other.Concurrency
		}
	}
	return mvc
}

// TopicConfig contains configuration of a pubsub topic
type TopicConfig struct {
	Name          string               `json:"name" yaml:"name"`
	BufferSize    int                  `json:"bufferSize,omitempty" yaml:"bufferSize,omitempty"`
	RateLimit     float64              `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	Scoring       *ScoringParams       `json:"scoring,omitempty" yaml:"scoring,omitempty"`
	MsgValidator  *MsgValidationConfig `json:"msgValidator,omitempty" yaml:"msgValidator,omitempty"`
	MsgIDFnConfig *MsgIDFnConfig       `json:"msgIDFn,omitempty" yaml:"msgIDFn,omitempty"`
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
	// whether it is allowed to just set some params and not all of them
	SkipAtomic bool `json:"skipAtomic,omitempty" yaml:"skipAtomic,omitempty"`
	// thresholds to use for scoring
	Thresholds *ScoringThresholds `json:"thresholds,omitempty" yaml:"thresholds,omitempty"`
	// params (P1-P4) per topic
	Topics map[string]TopicScoreParams `json:"topics,omitempty" yaml:"topics,omitempty"`
	// Aggregate topic score cap; this limits the total contribution of topics towards a positive
	// score. It must be positive (or 0 for no cap).
	TopicScoreCap float64 `json:"topicScoreCap,omitempty" yaml:"topicScoreCap,omitempty"`
	// P5: Application-specific peer scoring
	AppSpecific *AppSpecificScoring `json:"appSpecificScore,omitempty" yaml:"appSpecificScore,omitempty"`
	// P6: IP-colocation factor.
	// The parameter has an associated counter which counts the number of peers with the same IP.
	IPColocationFactor *IpColocationScoring `json:"ipColocationFactor,omitempty" yaml:"ipColocationFactor,omitempty"`
	// P7: behavioural pattern penalties.
	// This parameter has an associated counter which tracks misbehaviour as detected by the
	// router. The router currently applies penalties for the following behaviors:
	// - attempting to re-graft before the prune backoff time has elapsed.
	// - not following up in IWANT requests for messages advertised with IHAVE.
	BehaviourPenalty *ScoringParamExtended `json:"behaviourPenalty,omitempty" yaml:"behaviourPenalty,omitempty"`
}

type ScoringThresholds struct {
	// whether it is allowed to just set some params and not all of them.
	SkipAtomic bool `json:"skipAtomic,omitempty" yaml:"skipAtomic,omitempty"`
	// GossipThreshold is the score threshold below which gossip propagation is suppressed;
	// should be negative.
	GossipThreshold float64 `json:"gossip,omitempty" yaml:"gossip,omitempty"`
	// PublishThreshold is the score threshold below which we shouldn't publish when using flood
	// publishing (also applies to fanout and floodsub peers); should be negative and <= GossipThreshold.
	PublishThreshold float64 `json:"publish,omitempty" yaml:"publish,omitempty"`
	// GraylistThreshold is the score threshold below which message processing is suppressed altogether,
	// implementing an effective gray list according to peer score; should be negative and <= PublishThreshold.
	GraylistThreshold float64 `json:"graylist,omitempty" yaml:"graylist,omitempty"`
	// AcceptPXThreshold is the score threshold below which PX will be ignored; this should be positive
	// and limited to scores attainable by bootstrappers and other trusted nodes.
	AcceptPXThreshold float64 `json:"acceptPx,omitempty" yaml:"acceptPx,omitempty"`
	// OpportunisticGraftThreshold is the median mesh score threshold before triggering opportunistic
	// grafting; this should have a small positive value.
	OpportunisticGraftThreshold float64 `json:"opportunisticGraft,omitempty" yaml:"opportunisticGraft,omitempty"`
}

func (sc *ScoringParams) ToStd() (*pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	stdParams := pubsub.PeerScoreParams{
		SkipAtomicValidation: sc.SkipAtomic,
		Topics:               make(map[string]*pubsub.TopicScoreParams),
		TopicScoreCap:        sc.TopicScoreCap,
	}

	for topic, tsp := range sc.Topics {
		stdParams.Topics[topic] = tsp.ToStd()
	}

	if p := sc.AppSpecific; p != nil {
		stdParams.AppSpecificScore = p.ScoreFn
		stdParams.AppSpecificWeight = p.Weight
	}

	if p := sc.IPColocationFactor; p != nil {
		stdParams.IPColocationFactorThreshold = p.Threshold
		stdParams.IPColocationFactorWeight = p.Weight
		stdParams.IPColocationFactorWhitelist = p.Whitelist
	}

	if p := sc.BehaviourPenalty; p != nil {
		stdParams.BehaviourPenaltyWeight = p.Weight
		stdParams.BehaviourPenaltyDecay = p.Decay
		stdParams.BehaviourPenaltyThreshold = p.Threshold
	}

	stdThresholds := &pubsub.PeerScoreThresholds{}

	if sc.Thresholds != nil {
		stdThresholds.SkipAtomicValidation = sc.Thresholds.SkipAtomic
		stdThresholds.GossipThreshold = sc.Thresholds.GossipThreshold
		stdThresholds.PublishThreshold = sc.Thresholds.PublishThreshold
		stdThresholds.GraylistThreshold = sc.Thresholds.GraylistThreshold
		stdThresholds.AcceptPXThreshold = sc.Thresholds.AcceptPXThreshold
		stdThresholds.OpportunisticGraftThreshold = sc.Thresholds.OpportunisticGraftThreshold
	}

	return &stdParams, stdThresholds
}

// TopicScoreParams is used to configure topic scoring.
// The struct provides a simpler abstraction of pubsub.TopicScoreParams.
// for more information please refer to pubsub.TopicScoreParams.
type TopicScoreParams struct {
	// whether it is allowed to just set some params and not all of them
	SkipAtomic bool `json:"skipAtomic,omitempty" yaml:"skipAtomic,omitempty"`
	// The weight of the topic
	Weight float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
	// P1: time in the mesh
	// This is the time the peer has been grafted in the mesh.
	TimeInMesh *ScoringParamBase `json:"timeInMesh,omitempty" yaml:"timeInMesh,omitempty"`
	// P2: first message deliveries
	// This is the number of message deliveries in the topic.
	FirstMessageDeliveries *ScoringParamBase `json:"firstMessageDeliveries,omitempty" yaml:"firstMessageDeliveries,omitempty"`
	// P3: mesh message deliveries
	// This is the number of message deliveries in the mesh, within the MeshMessageDeliveriesWindow of
	// message validation; deliveries during validation also count and are retroactively applied
	// when validation succeeds.
	// This window accounts for the minimum time before a hostile mesh peer trying to game the score
	// could replay back a valid message we just sent them.
	// It effectively tracks first and near-first deliveries, i.e., a message seen from a mesh peer
	// before we have forwarded it to them.
	MeshMessageDeliveries *ScoringParamExtended `json:"meshMessageDeliveries,omitempty" yaml:"meshMessageDeliveries,omitempty"`
	// P3b: sticky mesh propagation failures
	// This is a sticky penalty that applies when a peer gets pruned from the mesh with an active
	// mesh message delivery penalty.
	MeshFailurePenalty *ScoringParamBase `json:"meshFailurePenalty,omitempty" yaml:"meshFailurePenalty,omitempty"`
	// P4: invalid messages
	// This is the number of invalid messages in the topic.
	InvalidMessageDeliveries *ScoringParamBase `json:"invalidMessageDeliveries,omitempty" yaml:"invalidMessageDeliveries,omitempty"`
}

// ToStd translates the TopicScoreParams to pubsub.TopicScoreParams
func (tsp *TopicScoreParams) ToStd() *pubsub.TopicScoreParams {
	stdParams := &pubsub.TopicScoreParams{
		SkipAtomicValidation: tsp.SkipAtomic,
		TopicWeight:          tsp.Weight,
	}

	if p := tsp.TimeInMesh; p != nil {
		stdParams.TimeInMeshCap = p.Cap
		stdParams.TimeInMeshQuantum = p.Quantum
		stdParams.TimeInMeshWeight = p.Weight
	}

	if p := tsp.FirstMessageDeliveries; p != nil {
		stdParams.FirstMessageDeliveriesCap = p.Cap
		stdParams.FirstMessageDeliveriesDecay = p.Decay
		stdParams.FirstMessageDeliveriesWeight = p.Weight
	}

	if p := tsp.MeshMessageDeliveries; p != nil {
		stdParams.MeshMessageDeliveriesCap = p.Cap
		stdParams.MeshMessageDeliveriesDecay = p.Decay
		stdParams.MeshMessageDeliveriesWeight = p.Weight
		stdParams.MeshMessageDeliveriesThreshold = p.Threshold
		stdParams.MeshMessageDeliveriesActivation = p.Activation
		stdParams.MeshMessageDeliveriesWindow = p.Window
	}

	if p := tsp.MeshFailurePenalty; p != nil {
		stdParams.MeshFailurePenaltyDecay = p.Decay
		stdParams.MeshFailurePenaltyWeight = p.Weight
	}

	if p := tsp.InvalidMessageDeliveries; p != nil {
		stdParams.InvalidMessageDeliveriesDecay = p.Decay
		stdParams.InvalidMessageDeliveriesWeight = p.Weight
	}

	return stdParams
}

type AppSpecificScoring struct {
	// TODO: think of way to expose this function through static config
	ScoreFn func(p peer.ID) float64
	Weight  float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
}

type ScoringParamBase struct {
	Weight float64 `json:"weight,omitempty" yaml:"weight,omitempty"`
	Decay  float64 `json:"decay,omitempty" yaml:"decay,omitempty"`
	Cap    float64 `json:"cap,omitempty" yaml:"cap,omitempty"`
	// used only for TimeInMesh (p1)
	Quantum time.Duration `json:"quantum,omitempty" yaml:"quantum,omitempty"`
}

type ScoringParamExtended struct {
	Weight     float64       `json:"weight,omitempty" yaml:"weight,omitempty"`
	Decay      float64       `json:"decay,omitempty" yaml:"decay,omitempty"`
	Cap        float64       `json:"cap,omitempty" yaml:"cap,omitempty"`
	Threshold  float64       `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	Window     time.Duration `json:"window,omitempty" yaml:"window,omitempty"`
	Activation time.Duration `json:"activation,omitempty" yaml:"activation,omitempty"`
}

// P6: IP-colocation factor.
// The parameter has an associated counter which counts the number of peers with the same IP.
// If the number of peers in the same IP exceeds IPColocationFactorThreshold, then the value
// is the square of the difference, ie (PeersInSameIP - IPColocationThreshold)^2.
// If the number of peers in the same IP is less than the threshold, then the value is 0.
// The weight of the parameter MUST be negative, unless you want to disable for testing.
//
// Note: In order to simulate many IPs in a managable manner when testing,
// you can set the weight to 0 thus disabling the IP colocation penalty.
type IpColocationScoring struct {
	Weight    float64
	Threshold int
	Whitelist []*net.IPNet
}
