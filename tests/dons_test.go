package tests

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	grpcapi "github.com/amirylm/p2pmq/api/grpc"
	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/core"
	logging "github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type donConfig struct {
	nodes           int
	dons            int
	reportsInterval time.Duration
	subscribed      []string
}

func TestCrossDONCommunication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 16
	donsCfg := map[string]donConfig{
		"auto": {
			nodes: 10,
			dons:  1,
			subscribed: []string{
				"func",
				"mercury",
			},
			reportsInterval: time.Second * 1,
		},
		"func": {
			nodes: 7,
			dons:  1,
			subscribed: []string{
				"auto",
				"mercury",
			},
			reportsInterval: time.Second * 1,
		},
		"mercury": {
			nodes:           4,
			dons:            1,
			reportsInterval: time.Second * 2,
		},
		"transmit": {
			nodes: 10,
			dons:  1,
			subscribed: []string{
				"func",
				"mercury",
			},
			reportsInterval: time.Second * 4,
		},
		"test": {
			nodes: 10,
			dons:  1,
			subscribed: []string{
				"auto",
				"func",
				"mercury",
			},
			reportsInterval: time.Second * 5,
		},
	}

	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "debug"))

	controllers, _, _, done := core.SetupTestControllers(ctx, t, n, func(*pubsub.Message) {}, func(peer.ID, *pubsub.Message) pubsub.ValidationResult {
		return pubsub.ValidationAccept
	})
	defer done()
	require.Equal(t, n, len(controllers))

	grpcServers := make([]*grpc.Server, n)
	for i := 0; i < n; i++ {
		ctrl := controllers[i]
		control, msgR, valR := grpcapi.NewServices(ctrl, 128)
		ctrl.RefreshRouters(func(mw *core.MsgWrapper[error]) {
			require.NoError(t, msgR.Push(mw))
		}, func(mw *core.MsgWrapper[pubsub.ValidationResult]) {
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			mw.Result = valR.PushWait(ctx, mw)
		})
		grpcServers[i] = grpcapi.NewGrpcServer(control, msgR, valR)
	}

	threadC := utils.NewThreadControl()
	defer threadC.Close()

	addrs := make([]string, n)
	nodes := make([]*nodeClient, n)
	for i, s := range grpcServers {
		{
			srv := s
			port := randPort()
			addrs[i] = fmt.Sprintf(":%d", port)
			nodes[i] = newNodeClient(fmt.Sprintf(":%d", port))
			threadC.Go(func(ctx context.Context) {
				err := grpcapi.ListenGrpc(srv, port)
				if ctx.Err() == nil {
					require.NoError(t, err)
				}
			})
		}
	}

	<-time.After(time.Second * 5) // TODO: avoid timeout

	dons := make(map[string][]*mockedDon)
	for did, cfg := range donsCfg {
		for j := 0; j < cfg.dons; j++ {
			donNodes := getRandomNodes(cfg.nodes, nodes)
			don := newDon(did, donNodes...)
			dons[did] = append(dons[did], don)
		}
	}
	for did, instances := range dons {
		cfg := donsCfg[did]
		for _, don := range instances {
			don.run(cfg.reportsInterval, cfg.subscribed...)
		}
	}
	defer func() {
		for _, instances := range dons {
			for _, don := range instances {
				don.stop()
			}
		}
	}()

	testDuration := time.Second * 20
	expectedReports := map[string]int{
		"auto":     int(testDuration) / int(donsCfg["auto"].reportsInterval),
		"func":     int(testDuration) / int(donsCfg["func"].reportsInterval),
		"mercury":  int(testDuration) / int(donsCfg["mercury"].reportsInterval),
		"transmit": int(testDuration) / int(donsCfg["transmit"].reportsInterval),
		"test":     int(testDuration) / int(donsCfg["test"].reportsInterval),
	}

	for ctx.Err() == nil {
		<-time.After(time.Second * 5)
		for did, exp := range expectedReports {
			instances := dons[did]
			for _, don := range instances {
				reportsCount := don.reportsCount()
				for reportsCount < exp && ctx.Err() == nil {
					time.Sleep(time.Second)
					reportsCount = don.reportsCount()
				}
			}
			if ctx.Err() == nil {
				t.Logf("DON %s reports count: %d", did, expectedReports[did])
				// we have enough reports for this don
				expectedReports[did] = 0
			}
		}
	}

	// <-time.After(testDuration + testDuration/4)

	for did, exp := range expectedReports {
		require.Equal(t, 0, exp, "DON %s reports count", did)
	}
}

func getRandomNodes(n int, items []*nodeClient) []*nodeClient {
	if n > len(items) {
		n = len(items)
	}
	visited := map[int]bool{}
	randoms := make([]*nodeClient, 0)
	for len(randoms) < n {
		r := rand.Intn(len(items))
		if visited[r] {
			continue
		}
		visited[r] = true
		randoms = append(randoms, items[r])
	}
	return randoms
}

func randPort() int {
	return 5001 + rand.Intn(3000) + rand.Intn(2000)
}
