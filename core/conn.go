package core

import (
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (d *Daemon) connect(pi peer.AddrInfo) {
	switch d.host.Network().Connectedness(pi.ID) {
	case libp2pnetwork.CannotConnect, libp2pnetwork.Connected:
		return
	default:
	}
	if d.ctx.Err() != nil {
		return
	}
	err := d.host.Connect(d.ctx, pi)
	if err != nil {
		if d.ctx.Err() != nil {
			return
		}
		d.lggr.Warnw("peer connection failed", "err", err, "peer", pi)
		return
	}
	d.lggr.Debugw("new peer connected", "peer", pi)
}
