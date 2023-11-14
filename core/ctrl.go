package core

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/commons/utils"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"go.uber.org/zap"
)

var lggr = logging.Logger("p2pmq")

type Controller struct {
	utils.StartStopOnce
	threadControl utils.ThreadControl

	lggr *zap.SugaredLogger

	cfg commons.Config

	host    host.Host
	dht     *dht.IpfsDHT
	mdnsSvc mdns.Service
	pubsub  *pubsub.PubSub

	psManager pubsubManager
	denylist  pubsub.Blacklist
	subFilter pubsub.SubscriptionFilter

	valRouter MsgRouter[pubsub.ValidationResult]
	msgRouter MsgRouter[error]
}

func NewController(
	ctx context.Context,
	cfg commons.Config,
	msgRouter MsgRouter[error],
	valRouter MsgRouter[pubsub.ValidationResult],
	lggrNS string,
) (*Controller, error) {
	d := &Controller{
		threadControl: utils.NewThreadControl(),
		lggr:          lggr.Named(lggrNS).Named("controller"),
		cfg:           cfg,
		valRouter:     valRouter,
		msgRouter:     msgRouter,
	}
	err := d.setup(ctx, cfg)

	return d, err
}

func (c *Controller) ID() string {
	return c.host.ID().String()
}

func (c *Controller) RefreshRouters(msgHandler func(*MsgWrapper[error]), valHandler func(*MsgWrapper[pubsub.ValidationResult])) {
	if c.valRouter != nil {
		c.valRouter.RefreshHandler(valHandler)
		c.threadControl.Go(c.valRouter.Start)
	}
	if c.msgRouter != nil {
		c.msgRouter.RefreshHandler(msgHandler)
		c.threadControl.Go(c.msgRouter.Start)
	}
}

func (c *Controller) Start(ctx context.Context) {
	c.StartOnce(func() {
		// d.lggr.Debugf("starting controller with host %s", d.host.ID())

		if c.msgRouter != nil {
			c.threadControl.Go(c.msgRouter.Start)
		}

		if c.valRouter != nil {
			c.threadControl.Go(c.valRouter.Start)
		}

		if c.cfg.Discovery != nil {
			_, bootstrappers, err := parseDiscoveryConfig(*c.cfg.Discovery)
			if err != nil {
				c.lggr.Panicf("failed to parse discovery config: %w", err)
			}
			c.lggr.Debugw("connecting to bootstrappers", "bootstrappers", bootstrappers, "raw", c.cfg.Discovery.Bootstrappers)
			for _, b := range bootstrappers {
				c.connect(b)
			}
			if err := c.dht.Bootstrap(ctx); err != nil {
				c.lggr.Panicf("failed to start discovery: %w", err)
			}
		}
		if c.mdnsSvc != nil {
			if err := c.mdnsSvc.Start(); err != nil {
				c.lggr.Errorf("failed to start mdns: %w", err)
			}
		}
	})
}

func (c *Controller) Close() {
	c.StopOnce(func() {
		c.lggr.Debugf("closing controller with host %s", c.host.ID())
		c.threadControl.Close()
		if c.dht != nil {
			if err := c.dht.Close(); err != nil {
				c.lggr.Errorf("failed to close DHT: %w", err)
			}
		}
		if c.mdnsSvc != nil {
			if err := c.mdnsSvc.Close(); err != nil {
				c.lggr.Errorf("failed to close mdns: %w", err)
			}
		}
		if err := c.host.Close(); err != nil {
			c.lggr.Errorf("failed to close host: %w", err)
		}
	})
}

func (c *Controller) setup(ctx context.Context, cfg commons.Config) (err error) {
	cn := &cfg
	cn.Defaults()
	cfg = *cn

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
		opts = append(opts, libp2p.Routing(c.dhtRoutingFactory(ctx, dhtOpts...)))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return err
	}
	c.host = h
	c.lggr.Infow("created libp2p host", "peerID", h.ID(), "addrs", h.Addrs())

	if len(cfg.MdnsTag) > 0 {
		c.setupMdnsDiscovery(ctx, h, cfg.MdnsTag)
	}

	if cfg.Pubsub != nil {
		err := c.setupPubsubRouter(ctx, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}
