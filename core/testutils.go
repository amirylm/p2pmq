package core

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core/gossip"
)

func SetupTestControllers(ctx context.Context, t *testing.T, n int, routingFn func(*pubsub.Message), valFn func(peer.ID, *pubsub.Message) pubsub.ValidationResult) ([]*Controller, []MsgRouter[error], []MsgRouter[pubsub.ValidationResult], func()) {
	bootAddr := "/ip4/127.0.0.1/tcp/5001"
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

	hitMap := map[string]*atomic.Int32{}
	for i := 0; i < n; i++ {
		hitMap[fmt.Sprintf("test-%d", i+1)] = &atomic.Int32{}
	}

	controllers := make([]*Controller, n)
	msgRouters := make([]MsgRouter[error], n)
	valRouters := make([]MsgRouter[pubsub.ValidationResult], n)
	for i := 0; i < n; i++ {
		cfg := commons.Config{
			ListenAddrs: []string{
				"/ip4/127.0.0.1/tcp/0",
			},
			// MdnsTag: "p2pmq/mdns/test",
			Discovery: &commons.DiscoveryConfig{
				Mode:           commons.ModeServer,
				ProtocolPrefix: "p2pmq/kad/test",
				Bootstrappers: []string{
					fmt.Sprintf("%s/p2p/%s", bootAddr, boot.host.ID()),
				},
			},
			Pubsub: &commons.PubsubConfig{
				Topics: []commons.TopicConfig{
					{
						Name:         "test-1",
						MsgValidator: &commons.MsgValidationConfig{},
					},
				},
			},
		}
		msgRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[error]) {
			routingFn(mw.Msg)
		}, gossip.MsgIDSha256(20))
		valRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[pubsub.ValidationResult]) {
			res := valFn(mw.Peer, mw.Msg)
			mw.Result = res
		}, gossip.MsgIDSha256(20))
		c, err := NewController(ctx, cfg, msgRouter, valRouter, fmt.Sprintf("peer-%d", i+1))
		require.NoError(t, err)
		controllers[i] = c
		msgRouters[i] = msgRouter
		valRouters[i] = valRouter
		t.Logf("created controller %d: %s", i+1, c.host.ID())
	}

	for i, c := range controllers {
		c.Start(ctx)
		t.Logf("started controller %d: %s", i+1, c.host.ID())
	}

	waitControllersConnected(n)

	return controllers, msgRouters, valRouters, func() {
		boot.Close()
		for _, c := range controllers {
			c.Close()
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
