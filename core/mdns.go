package core

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

func (c *Controller) setupMdnsDiscovery(ctx context.Context, host host.Host, serviceTag string) {
	c.lggr.Debugf("setting up local mdns for host %s", c.host.ID())

	notifee := &mdnsNotifee{
		ctx: ctx,
		q:   make(chan peer.AddrInfo, 32),
	}
	c.mdnsSvc = mdns.NewMdnsService(host, serviceTag, notifee)

	c.threadControl.Go(func(ctx context.Context) {
		defer notifee.Close()

		for {
			select {
			case p := <-notifee.q:
				c.connect(p)
			case <-ctx.Done():
				return
			}
		}
	})
}

type mdnsNotifee struct {
	ctx context.Context
	q   chan peer.AddrInfo
}

func (m *mdnsNotifee) Close() {
	close(m.q)
}

// HandlePeerFound implements mdns.Notifee
func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if m.ctx.Err() == nil {
		select {
		case m.q <- pi:
		default:
		}
	}
}
