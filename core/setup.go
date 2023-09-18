package core

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/commons"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

func (d *Daemon) setup(ctx context.Context, cfg commons.Config) (err error) {
	c := &cfg
	c.Defaults()
	cfg = *c

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
	}

	var sk crypto.PrivKey
	sk, cfg.PrivateKey, err = commons.GetOrGeneratePrivateKey(cfg.PrivateKey)
	if err != nil {
		return err
	}
	opts = append(opts, libp2p.Identity(sk))
	opts = append(opts, libp2p.WithDialTimeout(cfg.DialTimeout))
	opts = append(opts, libp2p.UserAgent(fmt.Sprintf("p2pmq/%s", cfg.UserAgent)))
	opts = append(opts, libp2p.Ping(!cfg.DisablePing))
	opts = append(opts, libp2p.PrivateNetwork(cfg.PSK))

	if cfg.ConnManager != nil {
		cfg.ConnManager.Defaults()
		cm, err := connmgr.NewConnManager(cfg.ConnManager.LowWaterMark,
			cfg.ConnManager.HighWaterMark,
			connmgr.WithGracePeriod(cfg.ConnManager.GracePeriod),
		)
		if err != nil {
			return fmt.Errorf("could not create conn manager: %w", err)
		}
		opts = append(opts, libp2p.ConnectionManager(cm))
	}

	if cfg.NatPortMap {
		opts = append(opts, libp2p.NATPortMap())
	}

	if cfg.AutoNat {
		opts = append(opts, libp2p.EnableNATService())
	}

	if cfg.Discovery != nil {
		cfg.Discovery.Defaults()
		dmode, bootstrappers, err := parseDiscoveryConfig(*cfg.Discovery)
		if err != nil {
			return err
		}
		dhtOpts := []dht.Option{
			dht.ProtocolPrefix(protocol.ID(cfg.Discovery.ProtocolPrefix)),
			dht.Mode(dmode),
		}
		// TBD: for custom validators for reports
		// for name, val := range validators {
		// 	dhtOpts = append(dhtOpts, dht.Option(dht.NamespacedValidator(name, val)))
		// }
		if len(bootstrappers) > 0 {
			dhtOpts = append(dhtOpts, dht.Option(dht.BootstrapPeers(bootstrappers...)))
		}
		opts = append(opts, libp2p.Routing(d.dhtRoutingFactory(ctx, dhtOpts...)))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return err
	}
	d.host = h
	d.lggr.Infow("created libp2p host", "peerID", h.ID(), "addrs", h.Addrs())

	if len(cfg.MdnsTag) > 0 {
		d.setupMdnsDiscovery(ctx, h, cfg.MdnsTag)
	}

	if cfg.Pubsub != nil {
		err := d.setupPubsubRouter(ctx, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Daemon) setupMdnsDiscovery(ctx context.Context, host host.Host, serviceTag string) {
	d.lggr.Debugf("setting up local mdns for host %s", d.host.ID())

	notifee := &mdnsNotifee{
		ctx: ctx,
		q:   make(chan peer.AddrInfo, 32),
	}
	d.mdnsSvc = mdns.NewMdnsService(host, serviceTag, notifee)

	go func() {
		defer notifee.Close()

		for {
			select {
			case p := <-notifee.q:
				d.connect(p)
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
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
