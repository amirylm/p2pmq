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
)

func TestDaemon_Sanity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 4
	rounds := 10

	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "debug"))

	hitMap := map[string]*atomic.Int32{}
	for i := 0; i < n; i++ {
		hitMap[fmt.Sprintf("test-%d", i+1)] = &atomic.Int32{}
	}

	daemons, done := setupDaemons(ctx, t, n, func(m *pubsub.Message) {
		hitMap[m.GetTopic()].Add(1)
		// lggr.Infow("got pubsub message", "topic", m.GetTopic(), "from", m.GetFrom(), "data", string(m.GetData()))
	})
	defer done()

	<-time.After(time.Second * 5) // TODO: avoid timeout

	t.Log("peers connected")

	var wg sync.WaitGroup
	for _, d := range daemons {
		wg.Add(1)
		go func(d *Daemon) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				require.NoError(t, d.Subscribe(ctx, fmt.Sprintf("test-%d", i+1)))
			}
		}(d)
	}

	wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for r := 0; r < rounds; r++ {
		for i, d := range daemons {
			wg.Add(1)
			go func(d *Daemon, r, i int) {
				defer wg.Done()
				require.NoError(t, d.Publish(ctx, fmt.Sprintf("test-%d", i+1), []byte(fmt.Sprintf("round-%d-test-data-%d", r+1, i+1))))
			}(d, r, i)
		}
	}

	wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for topic, counter := range hitMap {
		count := int(counter.Load()) / n // per node
		require.Equal(t, rounds, count, "should get %d messages on topic %s", rounds, topic)
	}
}

func setupDaemons(ctx context.Context, t *testing.T, n int, routingFn func(*pubsub.Message)) ([]*Daemon, func()) {
	bootAddr := "/ip4/127.0.0.1/tcp/5001"
	boot, err := NewDaemon(ctx, commons.Config{
		ListenAddrs: []string{
			bootAddr,
		},
		MdnsTag: "p2pmq/mdns/test",
		Discovery: &commons.DiscoveryConfig{
			Mode:           commons.ModeBootstrapper,
			ProtocolPrefix: "p2pmq/kad/test",
		},
	}, nil, "boot")
	require.NoError(t, err)
	t.Logf("created bootstrapper %s", boot.host.ID())
	go func() {
		_ = boot.Start(ctx)
	}()

	t.Logf("started bootstrapper %s", boot.host.ID())

	<-time.After(time.Second * 2)

	hitMap := map[string]*atomic.Int32{}
	for i := 0; i < n; i++ {
		hitMap[fmt.Sprintf("test-%d", i+1)] = &atomic.Int32{}
	}

	daemons := make([]*Daemon, n)
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

		d, err := NewDaemon(ctx, cfg, NewRouter(1024, 4, routingFn), fmt.Sprintf("peer-%d", i+1))
		require.NoError(t, err)
		daemons[i] = d
		t.Logf("created daemon %d: %s", i+1, d.host.ID())
	}

	for i, d := range daemons {
		require.NoError(t, d.Start(ctx))
		t.Logf("started daemon %d: %s", i+1, d.host.ID())
	}

	waitDaemonsConnected(n)

	return daemons, func() {
		_ = boot.Close()
		for _, d := range daemons {
			_ = d.Close()
		}
	}
}

func waitDaemonsConnected(n int, daemons ...*Daemon) {
	for _, d := range daemons {
		connected := make([]peer.ID, 0)
		for len(connected) < n {
			peers := d.host.Network().Peers()
			for _, pid := range peers {
				switch d.host.Network().Connectedness(pid) {
				case libp2pnetwork.Connected:
					connected = append(connected, pid)
				default:
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
