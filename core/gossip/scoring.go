package gossip

import (
	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func PeerScores(cfg commons.PubsubConfig) (*pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	peerScores := &pubsub.PeerScoreParams{}
	if cfg.Scoring != nil {
		peerScores = cfg.Scoring.ToStd()
	}
	return peerScores, &pubsub.PeerScoreThresholds{
		// TODO: using reasonable defaults, requires tuning
		GossipThreshold:   -10000,
		PublishThreshold:  -2000,
		GraylistThreshold: -400,
	}
}
