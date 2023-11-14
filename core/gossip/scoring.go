package gossip

import (
	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func PeerScores(cfg commons.PubsubConfig) (*pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	peerScores, thresholds := &pubsub.PeerScoreParams{}, &pubsub.PeerScoreThresholds{
		// TODO: using reasonable defaults, requires tuning
		SkipAtomicValidation: true,
		GossipThreshold:      -10000,
		PublishThreshold:     -2000,
		GraylistThreshold:    -400,
	}
	if cfg.Scoring != nil {
		s, t := cfg.Scoring.ToStd()
		if s != nil {
			peerScores = s
		}
		if t != nil {
			thresholds = t
		}
	}
	return peerScores, thresholds
}
