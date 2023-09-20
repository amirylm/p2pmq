package core

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	logging "github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core/gossip"
)

func TestController_Sanity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 4
	rounds := 10

	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "debug"))

	valHitMap := map[string]*atomic.Int32{}
	msgHitMap := map[string]*atomic.Int32{}
	for i := 0; i < n; i++ {
		topic := fmt.Sprintf("test-%d", i+1)
		msgHitMap[topic] = &atomic.Int32{}
		valHitMap[topic] = &atomic.Int32{}
	}

	controllers, done := setupControllers(ctx, t, n, func(msg *pubsub.Message) {
		msgHitMap[msg.GetTopic()].Add(1)
		// lggr.Infow("got pubsub message", "topic", m.GetTopic(), "from", m.GetFrom(), "data", string(m.GetData()))
	}, func(p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		valHitMap[msg.GetTopic()].Add(1)
		return pubsub.ValidationAccept
	})
	defer done()

	<-time.After(time.Second * 5) // TODO: avoid timeout

	t.Log("peers connected")

	var wg sync.WaitGroup
	for _, d := range controllers {
		wg.Add(1)
		go func(c *Controller) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				require.NoError(t, c.Subscribe(ctx, fmt.Sprintf("test-%d", i+1)))
			}
		}(d)
	}

	wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for r := 0; r < rounds; r++ {
		for i, d := range controllers {
			wg.Add(1)
			go func(c *Controller, r, i int) {
				defer wg.Done()
				require.NoError(t, c.Publish(ctx, fmt.Sprintf("test-%d", i+1), []byte(fmt.Sprintf("round-%d-test-data-%d", r+1, i+1))))
			}(d, r, i)
		}
	}

	wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for topic, counter := range msgHitMap {
		count := int(counter.Load()) / n // per node
		require.Equal(t, rounds, count, "should get %d messages on topic %s", rounds, topic)
	}
}

func setupControllers(ctx context.Context, t *testing.T, n int, routingFn func(*pubsub.Message), valFn func(peer.ID, *pubsub.Message) pubsub.ValidationResult) ([]*Controller, func()) {
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
			Pubsub: &commons.PubsubConfig{},
		}
		msgRouter := NewMsgRouter[error](1024, 4, func(mw *MsgWrapper[error]) {
			routingFn(mw.Msg)
		}, gossip.MsgIDSha256(20))
		valRouter := NewMsgRouter[pubsub.ValidationResult](1024, 4, func(mw *MsgWrapper[pubsub.ValidationResult]) {
			res := valFn(mw.Peer, mw.Msg)
			mw.Result = res
		}, gossip.MsgIDSha256(20))
		c, err := NewController(ctx, cfg, msgRouter, valRouter, fmt.Sprintf("peer-%d", i+1))
		require.NoError(t, err)
		controllers[i] = c
		t.Logf("created controller %d: %s", i+1, c.host.ID())
	}

	for i, c := range controllers {
		c.Start(ctx)
		t.Logf("started controller %d: %s", i+1, c.host.ID())
	}

	waitControllersConnected(n)

	return controllers, func() {
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
