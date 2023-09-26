package grpcapi

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	logging "github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/proto"
)

func TestGrpc_Network(t *testing.T) {
	// t.Skip()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 4
	rounds := 5

	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "debug"))

	valHitMap := map[string]*atomic.Int32{}
	msgHitMap := map[string]*atomic.Int32{}
	for i := 0; i < n; i++ {
		topic := fmt.Sprintf("test-%d", i+1)
		msgHitMap[topic] = &atomic.Int32{}
		valHitMap[topic] = &atomic.Int32{}
	}

	controllers, _, _, done := core.SetupTestControllers(ctx, t, n, func(msg *pubsub.Message) {
		msgHitMap[msg.GetTopic()].Add(1)
		// lggr.Infow("got pubsub message", "topic", m.GetTopic(), "from", m.GetFrom(), "data", string(m.GetData()))
	}, func(p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		valHitMap[msg.GetTopic()].Add(1)
		return pubsub.ValidationAccept
	})
	defer done()
	require.Equal(t, n, len(controllers))

	grpcServers := make([]*grpc.Server, n)
	for i := 0; i < n; i++ {
		ctrl := controllers[i]
		control, msgR, valR := NewServices(ctrl, 128)
		ctrl.RefreshRouters(func(mw *core.MsgWrapper[error]) {
			require.NoError(t, msgR.Push(mw))
		}, func(mw *core.MsgWrapper[pubsub.ValidationResult]) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			mw.Result = valR.PushWait(ctx, mw)
		})
		grpcServers[i] = NewGrpcServer(control, msgR, valR)
	}

	threadC := utils.NewThreadControl()
	defer threadC.Close()

	ports := make([]int, n)
	for i, s := range grpcServers {
		{
			srv := s
			port := randPort()
			ports[i] = port
			threadC.Go(func(ctx context.Context) {
				err := ListenGrpc(srv, port)
				if ctx.Err() == nil {
					require.NoError(t, err)
				}
			})
		}
	}

	<-time.After(time.Second * 5) // TODO: avoid timeout

	conns := make([]*grpc.ClientConn, n)
	for i := range grpcServers {
		conn, err := grpc.Dial(fmt.Sprintf(":%d", ports[i]), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		conns[i] = conn
	}

	for i := range grpcServers {
		conn, err := grpc.Dial(fmt.Sprintf(":%d", ports[i]), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		threadC.Go(func(ctx context.Context) {
			val := proto.NewValidationRouterClient(conn)
			valClient, err := val.Handle(ctx)
			require.NoError(t, err)

			for ctx.Err() == nil {
				msg, err := valClient.Recv()
				if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
					return
				}
				require.NoError(t, err)
				valHitMap[msg.GetTopic()].Add(1)
				if len(msg.Data) > 48 {
					require.NoError(t, valClient.Send(&proto.ValidatedMessage{
						Result: proto.ValidationResult_REJECT,
						Msg:    msg,
					}))
				} else if len(msg.Data) > 32 {
					require.NoError(t, valClient.Send(&proto.ValidatedMessage{
						Result: proto.ValidationResult_IGNORE,
						Msg:    msg,
					}))
				} else {
					require.NoError(t, valClient.Send(&proto.ValidatedMessage{
						Result: proto.ValidationResult_ACCEPT,
						Msg:    msg,
					}))
				}
			}
		})
	}

	for i := range grpcServers {
		conn, err := grpc.Dial(fmt.Sprintf(":%d", ports[i]), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		threadC.Go(func(ctx context.Context) {
			msgRouter := proto.NewMsgRouterClient(conn)
			client, err := msgRouter.Listen(ctx, &proto.ListenRequest{})
			require.NoError(t, err)

			for ctx.Err() == nil {
				msg, err := client.Recv()
				if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
					return
				}
				require.NoError(t, err)
				msgHitMap[msg.GetTopic()].Add(1)
				require.LessOrEqualf(t, len(msg.Data), 32, "should see only valid messages: %s", msg.Data)
			}
		})
	}

	var wg sync.WaitGroup
	for i := range grpcServers {
		control := proto.NewControlServiceClient(conns[i])
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				_, err := control.Subscribe(ctx, &proto.SubscribeRequest{
					Topic: fmt.Sprintf("test-%d", i+1),
				})
				require.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	<-time.After(time.Second * 5) // TODO: avoid timeout
	t.Log("Publishing")
	for r := 0; r < rounds; r++ {
		for i := range grpcServers {
			control := proto.NewControlServiceClient(conns[i])
			req := &proto.PublishRequest{
				Topic: fmt.Sprintf("test-%d", i+1),
				Data:  []byte(fmt.Sprintf("round-%d-test-data-%d", r+1, i+1)),
			}
			wg.Add(1)
			go func() {
				defer wg.Done()

				c, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()
				_, err := control.Publish(c, req)
				require.NoError(t, err)
			}()
		}
	}
	wg.Wait()

	// // invalid messages
	// for i := range grpcServers {
	// 	control := proto.NewControlServiceClient(conns[i])
	// 	data := []byte(fmt.Sprintf("%d-test-data-%d", rand.Int31n(1e3), i+1))
	// 	for len(data)+1 < 48 {
	// 		data = append(data, []byte(fmt.Sprintf("%d", 1e5+rand.Int31n(1e9)))...)
	// 	}
	// 	req := &proto.PublishRequest{
	// 		Topic: fmt.Sprintf("test-%d", i+1),
	// 		Data:  data,
	// 	}
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		_, _ = control.Publish(ctx, req)
	// 	}()
	// }

	// // ignored messages
	// for i := range grpcServers {
	// 	control := proto.NewControlServiceClient(conns[i])
	// 	data := []byte(fmt.Sprintf("%d-test-data-%d", rand.Int31n(1e3), i+1))
	// 	for len(data)+1 < 32 {
	// 		data = append(data, []byte(fmt.Sprintf("%d", rand.Int31n(1e3)))...)
	// 	}
	// 	req := &proto.PublishRequest{
	// 		Topic: fmt.Sprintf("test-%d", i+1),
	// 		Data:  data,
	// 	}
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		_, _ = control.Publish(ctx, req)
	// 	}()
	// }
	// wg.Wait()

	<-time.After(time.Second * 2) // TODO: avoid timeout

	for _, s := range grpcServers {
		s.Stop()
	}

	t.Log("Asserting")
	for topic, counter := range msgHitMap {
		count := int(counter.Load()) / n // per node
		require.GreaterOrEqual(t, count, rounds, "should get at least %d messages on topic %s", rounds, topic)
	}
}

func randPort() int {
	return 5000 + rand.Intn(5000)
}
