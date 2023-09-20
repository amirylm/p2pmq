package gossip

import (
	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func PeerScores(cfg commons.PubsubConfig) (*pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	return &pubsub.PeerScoreParams{
			// TODO
		}, &pubsub.PeerScoreThresholds{
			// TODO: using reasonable defaults, requires tuning
			GossipThreshold:   -10000,
			PublishThreshold:  -2000,
			GraylistThreshold: -400,
		}
}
