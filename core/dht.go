package core

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/commons"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

func (c *Controller) dhtRoutingFactory(ctx context.Context, opts ...dhtopts.Option) func(host.Host) (routing.PeerRouting, error) {
	return func(h host.Host) (routing.PeerRouting, error) {
		dhtInst, err := dht.New(ctx, h, opts...)
		if err != nil {
			return nil, err
		}
		c.dht = dhtInst
		return dhtInst, nil
	}
}

func parseDiscoveryConfig(opts commons.DiscoveryConfig) (dht.ModeOpt, []peer.AddrInfo, error) {
	var dmode dht.ModeOpt
	switch opts.Mode {
	case commons.ModeClient:
		dmode = dht.ModeClient
	case commons.ModeBootstrapper:
		dmode = dht.ModeServer
	case commons.ModeServer:
		dmode = dht.ModeAutoServer
	default:
		dmode = dht.ModeAuto
	}
	bootstrappers := make([]peer.AddrInfo, 0)
	for _, bnstr := range opts.Bootstrappers {
		bn, err := peer.AddrInfoFromString(bnstr)
		if err != nil {
			return dmode, nil, fmt.Errorf("failed to parse bootstrapper addr %s: %w", bnstr, err)
		}
		bootstrappers = append(bootstrappers, *bn)
	}

	return dmode, bootstrappers, nil
}
