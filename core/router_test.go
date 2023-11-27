package core

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/core/gossip"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestMsgRouter(t *testing.T) {
	msgHitMap := map[string]*atomic.Int32{}
	topicNameHandle := "test"
	topicNameHandleWait := "test2"

	msgHitMap[topicNameHandle] = &atomic.Int32{}
	msgHitMap[topicNameHandleWait] = &atomic.Int32{}

	mr := NewMsgRouter(1024, 4, func(msg *MsgWrapper[string]) {
		if msg == nil || msg.Msg == nil {
			return
		}
		topic := msg.Msg.GetTopic()
		msgHitMap[topic].Add(1)
		if topic == topicNameHandleWait {
			msg.Result = "ok"
		}
	}, gossip.DefaultMsgIDFn)
	tc := utils.NewThreadControl()
	tc.Go(mr.Start)
	defer tc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pid, err := peer.Decode("12D3KooWETtZ1bYX9KNZMfLNkGdzAMamfsrdvJjoG1gwvNynQ5za")
	require.NoError(t, err)

	t.Run("handle", func(t *testing.T) {
		numOfMsgs := 10

		for i := 0; i < numOfMsgs; i++ {
			require.NoError(t, mr.Handle(ctx, pid, &pubsub.Message{
				Message: &pubsub_pb.Message{
					Data:  []byte(fmt.Sprintf("dummy msg %d", i)),
					Topic: &topicNameHandle,
				},
			}))
		}

		tctx, tcancel := context.WithTimeout(ctx, 5*time.Second)
		defer tcancel()
		for tctx.Err() == nil {
			if msgHitMap[topicNameHandle].Load() == int32(numOfMsgs) {
				break
			}
		}
		require.Equal(t, int32(numOfMsgs), msgHitMap[topicNameHandle].Load())
	})

	t.Run("handle wait", func(t *testing.T) {
		numOfMsgs := 10

		for i := 0; i < numOfMsgs; i++ {
			res, err := mr.HandleWait(ctx, pid, &pubsub.Message{
				Message: &pubsub_pb.Message{
					Data:  []byte(fmt.Sprintf("dummy msg with wait %d", i)),
					Topic: &topicNameHandleWait,
				},
			})
			require.NoError(t, err)
			require.Equal(t, "ok", res)
		}

		tctx, tcancel := context.WithTimeout(ctx, 5*time.Second)
		defer tcancel()
		for tctx.Err() == nil {
			if msgHitMap[topicNameHandleWait].Load() == int32(numOfMsgs) {
				break
			}
		}
		require.Equal(t, int32(numOfMsgs), msgHitMap[topicNameHandleWait].Load())
	})
}
