package gossip

import (
	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// GossipSubParams returns the default pubsub.GossipSubParams parameters,
// overriden by the given overlay parameters.
func GossipSubParams(overlayParams *commons.OverlayParams) pubsub.GossipSubParams {
	gsCfg := pubsub.DefaultGossipSubParams()
	if overlayParams == nil {
		return gsCfg
	}
	if overlayParams.D > 0 {
		gsCfg.D = int(overlayParams.D)
	}
	if overlayParams.Dlow > 0 {
		gsCfg.Dlo = int(overlayParams.Dlow)
	}
	if overlayParams.Dhi > 0 {
		gsCfg.Dhi = int(overlayParams.Dhi)
	}
	if overlayParams.Dlazy > 0 {
		gsCfg.Dlazy = int(overlayParams.Dlazy)
	}
	if overlayParams.McacheGossip > 0 {
		gsCfg.MaxIHaveMessages = int(overlayParams.McacheGossip)
	}
	if overlayParams.McacheLen > 0 {
		gsCfg.MaxIHaveLength = int(overlayParams.McacheLen)
	}
	if fanoutTtl := overlayParams.FanoutTtl; fanoutTtl.Milliseconds() > 0 {
		gsCfg.FanoutTTL = fanoutTtl
	}
	if heartbeat := overlayParams.HeartbeatInterval; heartbeat.Milliseconds() > 0 {
		gsCfg.HeartbeatInterval = heartbeat
	}
	return gsCfg
}
