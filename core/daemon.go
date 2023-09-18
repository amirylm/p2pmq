package core

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/commons"
	logging "github.com/ipfs/go-log"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"go.uber.org/zap"
)

var lggr = logging.Logger("p2pmq")

type Daemon struct {
	lggr *zap.SugaredLogger

	ctx context.Context

	cfg commons.Config

	host    host.Host
	dht     *dht.IpfsDHT
	mdnsSvc mdns.Service
	pubsub  *pubsub.PubSub

	manager  pubsubManager
	denylist pubsub.Blacklist
	router   *Router[*pubsub.Message]
}

func NewDaemon(ctx context.Context, cfg commons.Config, router *Router[*pubsub.Message], lggrNS string) (*Daemon, error) {
	d := new(Daemon)

	d.ctx = ctx
	d.lggr = lggr.Named(lggrNS).Named("daemon")
	d.cfg = cfg
	rlggr := d.lggr.Named("router")
	if router == nil {
		router = NewRouter(1024, 4, func(m *pubsub.Message) {
			rlggr.Infow("got pubsub message", "topic", m.GetTopic(), "from", m.GetFrom(), "data", string(m.GetData()))
		})
	}
	d.router = router
	err := d.setup(ctx, cfg)

	return d, err
}

func (d *Daemon) Start(ctx context.Context) error {
	// d.lggr.Debugf("starting daemon with host %s", d.host.ID())

	go d.router.Start(ctx)

	if d.cfg.Discovery != nil {
		_, bootstrappers, err := parseDiscoveryConfig(*d.cfg.Discovery)
		if err != nil {
			return err
		}
		d.lggr.Debugw("connecting to bootstrappers", "bootstrappers", bootstrappers, "raw", d.cfg.Discovery.Bootstrappers)
		for _, b := range bootstrappers {
			d.connect(b)
		}
		if err := d.dht.Bootstrap(ctx); err != nil {
			return fmt.Errorf("failed to start discovery: %w", err)
		}
	}
	if d.mdnsSvc != nil {
		if err := d.mdnsSvc.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Daemon) Close() error {
	// d.lggr.Debugf("closing daemon with host %s", d.host.ID())
	if d.dht != nil {
		if err := d.dht.Close(); err != nil {
			return fmt.Errorf("failed to close DHT: %w", err)
		}
	}
	if d.mdnsSvc != nil {
		if err := d.mdnsSvc.Close(); err != nil {
			return fmt.Errorf("failed to close mdns: %w", err)
		}
	}
	return d.host.Close()
}
