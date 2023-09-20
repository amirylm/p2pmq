package core

import (
	"context"

	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (d *Controller) connect(pi peer.AddrInfo) {
	switch d.host.Network().Connectedness(pi.ID) {
	case libp2pnetwork.CannotConnect, libp2pnetwork.Connected:
		return
	default:
	}

	d.threadControl.Go(func(ctx context.Context) {
		err := d.host.Connect(ctx, pi)
		if err != nil {
			d.lggr.Warnw("peer connection failed", "err", err, "peer", pi)
			return
		}
		d.lggr.Debugw("new peer connected", "peer", pi)
	})
}
