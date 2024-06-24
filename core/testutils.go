package core

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core/gossip"
)

func SetupTestControllers(ctx context.Context, t *testing.T, n int, routingFn func(*pubsub.Message), valFn func(peer.ID, *pubsub.Message) pubsub.ValidationResult) ([]*Controller, []MsgRouter[error], []MsgRouter[pubsub.ValidationResult], func()) {
	bootAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 5000+rand.Intn(1000))
	boot, err := NewController(ctx, commons.Config{
		ListenAddrs: []string{
			bootAddr,
		},
		MdnsTag: "p2pmq/mdns/test",
		Discovery: &commons.DiscoveryConfig{
			Mode:           commons.ModeBootstrapper,
			ProtocolPrefix: "p2pmq/kad/test",
		},
	}, nil, nil, "boot")
	require.NoError(t, err)
	t.Logf("created bootstrapper %s", boot.host.ID())
	go boot.Start(ctx)

	t.Logf("started bootstrapper %s", boot.host.ID())

	<-time.After(time.Second * 2)

	gen := &DefaultGenerator{
		BootAddr:     fmt.Sprintf("%s/p2p/%s", bootAddr, boot.host.ID()),
		RoutingFn:    routingFn,
		ValidationFn: valFn,
	}
	controllers, msgRouters, valRouters, done, err := StartControllers(ctx, n, gen)
	require.NoError(t, err)

	t.Logf("created %d controllers", n)

	waitControllersConnected(n)

	return controllers, msgRouters, valRouters, func() {
		done()
		boot.Close()
	}
}

func StartControllers(ctx context.Context, n int, gen Generator) ([]*Controller, []MsgRouter[error], []MsgRouter[pubsub.ValidationResult], func(), error) {
	controllers := make([]*Controller, 0, n)
	msgRouters := make([]MsgRouter[error], 0, n)
	valRouters := make([]MsgRouter[pubsub.ValidationResult], 0, n)
	done := func() {
		for _, c := range controllers {
			c.Close()
		}
	}
	for i := 0; i < n; i++ {
		cfg, msgRouter, valRouter, name := gen.NextConfig(i)
		c, err := NewController(ctx, cfg, msgRouter, valRouter, name)
		if err != nil {
			return controllers, msgRouters, valRouters, done, err
		}
		controllers = append(controllers, c)
		msgRouters = append(msgRouters, msgRouter)
		valRouters = append(valRouters, valRouter)
	}

	for _, c := range controllers {
		c.Start(ctx)
	}

	return controllers, msgRouters, valRouters, done, nil
}

func waitMinConnected(ctx context.Context, minConnected func(i int) int, backoff time.Duration, hosts ...host.Host) {
	for i, h := range hosts {
		connected := make([]peer.ID, 0)
		min := minConnected(i)
		for len(connected) < min && ctx.Err() == nil {
			peers := h.Network().Peers()
			for _, pid := range peers {
				switch h.Network().Connectedness(pid) {
				case libp2pnetwork.Connected:
					connected = append(connected, pid)
				default:
				}
			}
			if len(connected) < min {
				fmt.Printf("host %s connected to %d peers, waiting for %d\n", h.ID(), len(connected), min)
			}
			time.Sleep(backoff)
		}
	}
}

func waitControllersConnected(n int, controllers ...*Controller) {
	for _, c := range controllers {
		connected := make([]peer.ID, 0)
		for len(connected) < n {
			peers := c.host.Network().Peers()
			for _, pid := range peers {
				switch c.host.Network().Connectedness(pid) {
				case libp2pnetwork.Connected:
					connected = append(connected, pid)
				default:
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

type Generator interface {
	NextConfig(i int) (commons.Config, MsgRouter[error], MsgRouter[pubsub.ValidationResult], string)
}

type DefaultGenerator struct {
	BootAddr     string
	RoutingFn    func(*pubsub.Message)
	ValidationFn func(peer.ID, *pubsub.Message) pubsub.ValidationResult
}

func (g *DefaultGenerator) NextConfig(i int) (commons.Config, MsgRouter[error], MsgRouter[pubsub.ValidationResult], string) {
	cfg := commons.Config{
		ListenAddrs: []string{
			"/ip4/127.0.0.1/tcp/0",
		},
		MdnsTag: "p2pmq/mdns/test",
		// Discovery: &commons.DiscoveryConfig{
		// 	Mode:           commons.ModeServer,
		// 	ProtocolPrefix: "p2pmq/kad/test",
		// 	Bootstrappers: []string{
		// 		g.BootAddr,
		// 		// fmt.Sprintf("%s/p2p/%s", g.BootAddr, boot.host.ID()),
		// 	},
		// },
		Pubsub: &commons.PubsubConfig{
			MsgValidator: &commons.MsgValidationConfig{},
		},
	}

	msgRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[error]) {
		g.RoutingFn(mw.Msg)
	}, gossip.DefaultMsgIDFn)

	valRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[pubsub.ValidationResult]) {
		res := g.ValidationFn(mw.Peer, mw.Msg)
		mw.Result = res
	}, gossip.DefaultMsgIDFn)

	return cfg, msgRouter, valRouter, fmt.Sprintf("node-%d", i+1)
}
