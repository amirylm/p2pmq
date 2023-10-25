package blstest

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/herumi/bls-eth-go-binary/bls"
	logging "github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	grpcapi "github.com/amirylm/p2pmq/api/grpc"
	"github.com/amirylm/p2pmq/api/grpc/client"
	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/core"
)

func TestBls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 10

	netCfg := map[string]netConfig{
		"auto": {
			nodes: 4,
			subscribed: []string{
				"func",
				"mercury",
			},
			reportsInterval: time.Second * 1,
		},
		"func": {
			nodes: 4,
			subscribed: []string{
				"auto",
				"mercury",
			},
			reportsInterval: time.Second * 1,
		},
		"mercury": {
			nodes: 4,
			subscribed: []string{
				"test",
			},
			reportsInterval: time.Second * 2,
		},
		"transmit": {
			nodes: 4,
			subscribed: []string{
				"func",
				"mercury",
			},
			reportsInterval: time.Second * 4,
		},
		"test": {
			nodes: 4,
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
		return pubsub.ValidationIgnore
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

	pubstore := NewStore[*bls.PublicKey]()
	blsKeys := make([]*bls.SecretKey, 0)
	for net := range netCfg {
		netPriv, netPub := GenBlsKey()
		pubstore.Add(net, netPub)
		blsKeys = append(blsKeys, netPriv)
	}

	t.Log("Starting grpc servers")
	addrs := make([]string, n)
	nodes := make([]*Node, n)
	for i, s := range grpcServers {
		{
			srv := s
			port := randPort()
			addrs[i] = fmt.Sprintf(":%d", port)
			nodes[i] = NewNode(client.GrpcEndPoint(fmt.Sprintf(":%d", port)), pubstore)
			go func() {
				err := grpcapi.ListenGrpc(srv, port)
				if ctx.Err() == nil {
					require.NoError(t, err)
				}
			}()
		}
	}
	defer func() {
		for _, s := range grpcServers {
			s.Stop()
		}
		t.Log("grpc servers stopped")
	}()

	t.Log("Starting nodes")
	for _, n := range nodes {
		n.Start()
	}
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	netInstances := make(map[string][]*Node)
	j := 0
	for net, cfg := range netCfg {
		sks, pks, err := GenShares(blsKeys[j], uint64(Threshold(cfg.nodes)), uint64(cfg.nodes))
		require.NoError(t, err)
		nodes := getRandomNodes(cfg.nodes, nodes)
		netInstances[net] = nodes
		for i, n := range nodes {
			n.Shares.Add(net, Share{
				Signers:        pks,
				SignerID:       uint64(i + 1),
				PrivateKey:     sks[uint64(i+1)],
				SharePublicKey: blsKeys[j].GetPublicKey(),
			})
			require.NoError(t, n.consumer.Subscribe(ctx, net))
			for _, sub := range cfg.subscribed {
				require.NoError(t, n.consumer.Subscribe(ctx, sub))
			}
		}
		j++
	}

	t.Log("Nodes subscribed")

	<-time.After(time.Second * 5) // TODO: avoid timeout

	t.Log("Starting reports generation")

	threadC := utils.NewThreadControl()
	defer threadC.Close()
	reports := NewReportBuffer(reportBufferSize)

	for net, cfg := range netCfg {
		nodes := netInstances[net]
		threadC.Go(func(ctx context.Context) {
			triggerReports(ctx, t, net, cfg.reportsInterval, reports, nodes)
		})
	}

	testDuration := time.Second * 20
	expectedReports := map[string]int{
		"auto":     int(testDuration) / int(netCfg["auto"].reportsInterval),
		"func":     int(testDuration) / int(netCfg["func"].reportsInterval),
		"mercury":  int(testDuration) / int(netCfg["mercury"].reportsInterval),
		"transmit": int(testDuration) / int(netCfg["transmit"].reportsInterval),
		"test":     int(testDuration) / int(netCfg["test"].reportsInterval),
	}

checkLoop:
	for ctx.Err() == nil {
		<-time.After(testDuration / 4)
		for did, exp := range expectedReports {
			reportsCount := len(reports.GetAll(did))
			for reportsCount+1 < exp && ctx.Err() == nil {
				time.Sleep(time.Second)
				reportsCount = len(reports.GetAll(did))
			}
			if ctx.Err() == nil {
				t.Logf("DON %s reports count: %d", did, expectedReports[did])
				// we have enough reports for this don
				expectedReports[did] = 0
			}
		}
		for _, exp := range expectedReports {
			if exp != 0 {
				continue checkLoop
			}
		}
		break
	}

	<-time.After(testDuration / 4)

	for did, exp := range expectedReports {
		require.Equal(t, 0, exp, "DON %s reports count", did)
	}
	t.Log("done")
	cancel()
	done()
}

func triggerReports(ctx context.Context, t *testing.T, net string, interval time.Duration, reports *ReportBuffer, nodes []*Node) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest := reports.getLatest(net)
			var nextSeq uint64
			if latest != nil {
				nextSeq = latest.SeqNumber + 1
			} else {
				nextSeq = 1
			}
			t.Logf("Generating report for %s, seq %d", net, nextSeq)
			for _, n := range nodes {
				node := n
				share, ok := node.Shares.Get(net)
				if !ok {
					return
				}
				if node.getLeader(net, nextSeq) == share.SignerID {
					node.threadC.Go(func(ctx context.Context) {
						report := &SignedReport{
							Network:   net,
							SeqNumber: nextSeq,
							Data:      []byte(fmt.Sprintf("report for %s, seq %d", net, nextSeq)),
						}
						share.Sign(report)
						require.NoError(t, node.Broadcast(ctx, *report))
						reports.Add(net, *report)
					})
				}
			}
		}
	}
}

type netConfig struct {
	nodes           int
	reportsInterval time.Duration
	subscribed      []string
}

func randPort() int {
	return 5001 + rand.Intn(3000) + rand.Intn(2000)
}

func getRandomNodes(n int, items []*Node) []*Node {
	if n > len(items) {
		n = len(items)
	}
	visited := map[int]bool{}
	randoms := make([]*Node, 0)
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
