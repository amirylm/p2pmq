package gossip

import (
	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func GossipSubParams(overlayp *commons.OverlayParams) pubsub.GossipSubParams {
	gsCfg := pubsub.DefaultGossipSubParams()
	var overlay commons.OverlayParams
	if overlayp != nil {
		overlay = *overlayp
	}
	if overlay.D > 0 {
		gsCfg.D = int(overlay.D)
	}
	if overlay.Dlow > 0 {
		gsCfg.Dlo = int(overlay.Dlow)
	}
	if overlay.Dhi > 0 {
		gsCfg.Dhi = int(overlay.Dhi)
	}
	if overlay.Dlazy > 0 {
		gsCfg.Dlazy = int(overlay.Dlazy)
	}
	if overlay.McacheGossip > 0 {
		gsCfg.MaxIHaveMessages = int(overlay.McacheGossip)
	}
	if overlay.McacheLen > 0 {
		gsCfg.MaxIHaveLength = int(overlay.McacheLen)
	}
	if fanoutTtl := overlay.FanoutTtl; fanoutTtl.Milliseconds() > 0 {
		gsCfg.FanoutTTL = fanoutTtl
	}
	if heartbeat := overlay.HeartbeatInterval; heartbeat.Milliseconds() > 0 {
		gsCfg.HeartbeatInterval = heartbeat
	}
	return gsCfg
}
