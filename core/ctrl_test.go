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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
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

	controllers, _, _, done := SetupTestControllers(ctx, t, n, func(msg *pubsub.Message) {
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
				require.NoError(t, c.Publish(ctx, fmt.Sprintf("test-%d", i+1), []byte(fmt.Sprintf("r-%d-test-data-%d", r+1, i+1))))
			}(d, r, i)
		}
	}

	wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for topic, counter := range msgHitMap {
		count := int(counter.Load()) / n // per node
		// add 1 to account for the first message sent by the node
		count += 1
		require.GreaterOrEqual(t, count, rounds, "should get %d messages on topic %s", rounds, topic)
	}
}
