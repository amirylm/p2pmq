package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"text/template"
	"time"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/core/gossip"
	logging "github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestGossipSimulation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "error"))

	groupsCfgSimple18 := testGroupSimple(18, 10, 6, 2)
	groupsCfgSimple36 := testGroupSimple(36, 16, 16, 4)
	groupsCfgSimple54 := testGroupSimple(54, 32, 16, 6)

	benchFlows := []flow{
		flowActionA2B(1, time.Millisecond*1, time.Millisecond*250),
		flowActionA2B(10, time.Millisecond*1, time.Millisecond*10),
		flowActionA2B(100, time.Millisecond*1, time.Millisecond*10),
		flowActionA2B(1000, time.Millisecond*10, time.Millisecond*10),
		flowTrigger("b", 1, time.Millisecond*10),
		flowTrigger("b", 10, time.Millisecond*10),
		flowTrigger("b", 100, time.Millisecond*10),
		flowTrigger("b", 1000, time.Millisecond*10),
		flowTrigger("a", 1, time.Millisecond*10),
		flowTrigger("a", 10, time.Millisecond*10),
		flowTrigger("a", 100, time.Millisecond*10),
		flowTrigger("a", 1000, time.Millisecond*10),
	}

	tests := []struct {
		name         string
		n            int
		pubsubConfig *commons.PubsubConfig
		gen          *testGen
		groupsCfg    groupsCfg
		conns        connectivity
		flows        []flow
	}{
		{
			name: "simple_18",
			n:    18,
			gen: &testGen{
				// hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace:        &commons.PubsubTraceConfig{},
					Overlay: &commons.OverlayParams{
						D:     3,
						Dlow:  2,
						Dhi:   5,
						Dlazy: 3,
					},
				},
			},
			groupsCfg: groupsCfgSimple18,
			conns:     groupsCfgSimple18.allToAllConnectivity(),
			flows:     benchFlows[:],
		},
		{
			name: "simple_36",
			n:    36,
			gen: &testGen{
				// hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace:        &commons.PubsubTraceConfig{},
					Overlay: &commons.OverlayParams{
						D:     4,
						Dlow:  2,
						Dhi:   6,
						Dlazy: 3,
					},
				},
			},
			groupsCfg: groupsCfgSimple36,
			conns:     groupsCfgSimple36.allToAllConnectivity(),
			flows:     benchFlows[:],
		},
		{
			name: "simple_36_with_default_overlay_params",
			n:    36,
			gen: &testGen{
				// hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace:        &commons.PubsubTraceConfig{},
				},
			},
			groupsCfg: groupsCfgSimple36,
			conns:     groupsCfgSimple36.allToAllConnectivity(),
			flows:     benchFlows[:],
		},
		{
			name: "simple_54_with_default_overlay_params",
			n:    54,
			gen: &testGen{
				// hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace:        &commons.PubsubTraceConfig{},
				},
			},
			groupsCfg: groupsCfgSimple54,
			conns:     groupsCfgSimple54.allToAllConnectivity(),
			flows:     benchFlows[:],
		},
	}

	outDir := os.Getenv("GOSSIP_OUT_DIR")
	if len(outDir) == 0 {
		outDir = t.TempDir()
	}
	gossipSimulation := os.Getenv("GOSSIP_SIMULATION")
	switch gossipSimulation {
	case "full":
		t.Log("running full simulation")
	default:
		t.Log("running limited simulation (only for 18 and 36 nodes and flows with less than 100 iterations)")
		tests = tests[:3]
		for i, tc := range tests {
			newFlows := make([]flow, 0, len(tc.flows))
			for _, f := range tc.flows {
				if f.iterations < 100 {
					newFlows = append(newFlows, f)
				}
			}
			tc.flows = newFlows
			tests[i] = tc
		}
	}

	var outputs []gossipTestOutput
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tgen := tc.gen
			ctrls, _, _, done, err := StartControllers(ctx, tc.n, tgen)
			defer done()
			require.NoError(t, err)

			t.Logf("%d controllers were started, connecting peers...", tc.n)

			require.NoError(t, tc.conns.connect(ctx, ctrls))

			<-time.After(time.Second * 2) // TODO: avoid timeout

			t.Log("subscribing to topics...")
			groups := make(map[string][]*Controller)
			for name, groupCfg := range tc.groupsCfg {
				for _, i := range groupCfg.ids {
					if i >= tc.n {
						continue
					}
					ctrl := ctrls[i]
					groups[name] = append(groups[name], ctrl)
					for _, topic := range groupCfg.subs {
						require.NoError(t, ctrl.Subscribe(ctx, topic))
					}
					for _, topic := range groupCfg.relays {
						require.NoError(t, ctrl.Relay(topic))
					}
				}
			}

			// waiting for nodes to setup subscriptions
			<-time.After(time.Second * 4) // TODO: avoid timeout

			// starting fresh trace after subscriptions
			startupFaucets := traceFaucets{}
			pubsubRpcCount := new(atomic.Uint64)
			for _, ctrl := range ctrls {
				startupFaucets.add(ctrl.psTracer.faucets)
				ctrl.psTracer.Reset()
				pubsubRpcCount.Store(ctrl.pubsubRpcCounter.Swap(0))
			}
			// t.Logf("\n [%s] startup in_rpc_count: %d\n; trace faucets: %+v\n", tc.name, startupFaucets, pubsubRpcCount.Load())

			d := uint(6)
			if tgen.pubsubConfig != nil && tgen.pubsubConfig.Overlay != nil {
				d = uint(tgen.pubsubConfig.Overlay.D)
			}
			baseTestOutput := gossipTestOutput{
				Name: tc.name,
				N:    uint(tc.n),
				A:    uint(len(groups["a"])),
				B:    uint(len(groups["b"])),
				R:    uint(len(groups["relayers"])),
				D:    uint(d),
			}

			startupTestOutput := baseTestOutput
			startupTestOutput.Iterations = 1
			startupTestOutput.InboundRPC = pubsubRpcCount.Load()
			startupTestOutput.Name = fmt.Sprintf("%s-startup", tc.name)
			outputs = append(outputs, startupTestOutput)

			t.Log("starting flows...")

			for _, f := range tc.flows {
				flow := f
				flowTestName := fmt.Sprintf("%s-x%d", flow.name, flow.iterations)
				t.Run(flowTestName, func(t *testing.T) {
					threadCtrl := utils.NewThreadControl()
					defer threadCtrl.Close()

					start := time.Now()

					for i := 0; i < flow.iterations; i++ {
						for _, e := range flow.events {
							event := e
							group, ok := groups[event.srcGroup]
							if !ok || len(group) == 0 {
								continue
							}
							var wg sync.WaitGroup
							for _, c := range group {
								_i := i
								ctrl := c
								ctrlName := strings.Replace(ctrl.lggr.Desugar().Name(), ".ctrl", "", 1)
								ctrlName = strings.Replace(ctrlName, "p2pmq.", "", 1)
								wg.Add(1)
								threadCtrl.Go(func(ctx context.Context) {
									defer wg.Done()
									msg := event.Msg(msgArgs{
										i:     _i,
										group: event.srcGroup,
										ctrl:  ctrlName,
										flow:  flow.name,
									})
									require.NoError(t, ctrl.Publish(ctx, event.topic, []byte(msg)))
									// msgID := gossip.DefaultMsgIDFn(&pubsub_pb.Message{Data: []byte(msg)})
									// hmap.addSent(msgID)
								})
							}
							if event.wait {
								wg.Wait()
							}
							if event.interval > 0 {
								<-time.After(event.interval)
							}
						}
						<-time.After(flow.interval)
					}

					<-time.After(time.Second * 2) // TODO: avoid timeout

					faucets := traceFaucets{}
					var pubsubRpcCount uint64
					for _, ctrl := range ctrls {
						nodeFaucets := ctrl.psTracer.faucets
						// t.Logf("[%s] trace faucets: %+v", ctrl.lggr.Desugar().Name(), nodeFaucets)
						faucets.add(nodeFaucets)
						pubsubRpcCount += ctrl.pubsubRpcCounter.Swap(0)
						ctrl.psTracer.Reset()
					}
					testOutput := baseTestOutput
					testOutput.Name = flow.name
					testOutput.Iterations = uint(flow.iterations)
					testOutput.Faucets = msgTraceFaucetsOutput{
						Publish: faucets.publish,
						Deliver: faucets.deliver,
						Reject:  faucets.reject,
						DropRPC: faucets.dropRPC,
						SendRPC: faucets.sendRPC,
						RecvRPC: faucets.recvRPC,
					}
					testOutput.InboundRPC = pubsubRpcCount
					testOutput.TotalTime = time.Since(start)
					outputs = append(outputs, testOutput)
					t.Logf("output: %+v", testOutput)
				})
			}
		})
	}
	outputsJson, err := json.Marshal(outputs)
	require.NoError(t, err)
	outputFileName := fmt.Sprintf("%s/test-%d.json", outDir, time.Now().UnixMilli())
	require.NoError(t, os.WriteFile(outputFileName, outputsJson, 0644))
	t.Logf("outputs saved in %s", outputFileName)
}

type connectivity map[int][]int

func (connect connectivity) connect(ctx context.Context, ctrls []*Controller) error {
	for i, ctrl := range ctrls {
		conns := connect[i]
		for _, c := range conns {
			if i == c {
				continue
			}
			if err := ctrl.Connect(ctx, ctrls[c]); err != nil {
				return err
			}
		}
	}
	return nil
}

type msgTraceFaucetsOutput struct {
	Publish int `json:"publish,omitempty"`
	Deliver int `json:"deliver,omitempty"`
	Reject  int `json:"reject,omitempty"`
	DropRPC int `json:"drop_rpc,omitempty"`
	SendRPC int `json:"send_rpc,omitempty"`
	RecvRPC int `json:"recv_rpc,omitempty"`
}

type gossipTestOutput struct {
	Name          string
	N, A, B, R, D uint
	Iterations    uint
	Faucets       msgTraceFaucetsOutput
	InboundRPC    uint64
	TotalTime     time.Duration
}

type groupsCfg map[string]groupCfg

// baseConnectivity returns a base connectivity map for the groups,
// where each group member is connected to all other members of the group
// and to the relayers.
// func (groups groupsCfg) baseConnectivity() connectivity {
// 	conns := make(connectivity)
// 	var relayerIDs []int
// 	relayers, ok := groups["relayers"]
// 	if ok {
// 		relayerIDs = relayers.ids
// 	}
// 	for _, cfg := range groups {
// 		connectIDs := append(cfg.ids, relayerIDs...)
// 		for _, i := range cfg.ids {
// 			conns[i] = connectIDs
// 		}
// 	}
// 	return conns
// }

func (groups groupsCfg) allToAllConnectivity() connectivity {
	conns := make(connectivity)
	var allIDs []int
	for _, cfg := range groups {
		allIDs = append(allIDs, cfg.ids...)
	}
	for _, id := range allIDs {
		conns[id] = allIDs
	}
	return conns
}

type groupCfg struct {
	ids    []int
	subs   []string
	relays []string
}

// testGroupSimple creates a simple test group configuration with n nodes:
// group a: a nodes
// group b: b nodes
// relayers: r nodes
// NOTE: n >= a + b + r must hold
func testGroupSimple(n, a, b, r int) groupsCfg {
	ids := make([]int, n)
	for i := 0; i < n; i++ {
		ids[i] = i
	}
	return groupsCfg{
		"a": {
			ids:  ids[:a],
			subs: []string{"b.action.res", "b.trigger"},
		},
		"b": {
			ids:  ids[a : a+b],
			subs: []string{"a.trigger", "b.action.req"},
		},
		"relayers": {
			ids:    ids[a+b : a+b+r],
			relays: []string{"a.trigger", "b.action.req", "b.action.res", "b.trigger"},
		},
	}
}

func flowActionA2B(iterations int, interval, waitAfterReq time.Duration) flow {
	return flow{
		name: "action a->b",
		events: []flowEvent{
			{
				srcGroup: "a",
				topic:    "b.action.req",
				pattern:  "dummy-request-{{.group}}-{{.i}}",
				interval: waitAfterReq,
				wait:     true,
			},
			{
				srcGroup: "b",
				topic:    "b.action.res",
				pattern:  "dummy-response-{{.group}}-{{.i}}",
			},
		},
		iterations: iterations,
		interval:   interval,
	}
}

func flowTrigger(src string, iterations int, interval time.Duration) flow {
	return flow{
		name: fmt.Sprintf("trigger %s", src),
		events: []flowEvent{
			{
				srcGroup: src,
				topic:    fmt.Sprintf("%s.trigger", src),
				pattern:  "dummy-trigger-{{.group}}-{{.i}}",
			},
		},
		iterations: iterations,
		interval:   interval,
	}
}

type flowEvent struct {
	srcGroup string
	topic    string
	pattern  string
	interval time.Duration
	wait     bool
}

type msgArgs struct {
	i     int
	group string
	ctrl  string
	flow  string
}

func (fe flowEvent) Msg(args msgArgs) string {
	tmpl, err := template.New("msg").Parse(fe.pattern)
	if err != nil {
		return ""
	}
	sb := new(strings.Builder)
	if err := tmpl.Execute(sb, map[string]interface{}{
		"i":     args.i,
		"group": args.group,
		"ctrl":  args.ctrl,
		"flow":  args.flow,
	}); err != nil {
		return ""
	}
	return sb.String()
}

type flow struct {
	name       string
	events     []flowEvent
	iterations int
	interval   time.Duration
}

// type nodeHitMap struct {
// 	lock           sync.RWMutex
// 	valHitMap      map[string]uint32
// 	msgHitMap      map[string]uint32
// 	sent, recieved map[string]time.Time
// }

// func (n *nodeHitMap) validations(topic string) uint32 {
// 	n.lock.RLock()
// 	defer n.lock.RUnlock()

// 	return n.valHitMap[topic]
// }

// func (n *nodeHitMap) messages(topic string) uint32 {
// 	n.lock.RLock()
// 	defer n.lock.RUnlock()

// 	return n.msgHitMap[topic]
// }

// func (n *nodeHitMap) addValidation(topic string) {
// 	n.lock.Lock()
// 	defer n.lock.Unlock()

// 	n.valHitMap[topic] += 1
// }

// func (n *nodeHitMap) addMessage(topic, msgID string) {
// 	n.lock.Lock()
// 	defer n.lock.Unlock()

// 	n.msgHitMap[topic] += 1
// 	n.recieved[msgID] = time.Now()
// }

// func (n *nodeHitMap) addSent(msgID string) {
// 	n.lock.Lock()
// 	defer n.lock.Unlock()

// 	n.sent[msgID] = time.Now()
// }

type testGen struct {
	// hitMaps      map[string]*nodeHitMap
	routingFn    func(*pubsub.Message)
	validationFn func(peer.ID, *pubsub.Message) pubsub.ValidationResult
	pubsubConfig *commons.PubsubConfig
}

func (g *testGen) NextConfig(i int) (commons.Config, MsgRouter[error], MsgRouter[pubsub.ValidationResult], string) {
	cfg := commons.Config{
		ListenAddrs: []string{
			"/ip4/127.0.0.1/tcp/0",
		},
		Pubsub: g.pubsubConfig,
	}

	name := fmt.Sprintf("node-%d", i+1)

	// hitMap := &nodeHitMap{
	// 	valHitMap: make(map[string]uint32),
	// 	msgHitMap: make(map[string]uint32),
	// 	recieved:  make(map[string]time.Time),
	// 	sent:      make(map[string]time.Time),
	// }
	// g.hitMaps[name] = hitMap

	msgRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[error]) {
		// hitMap.addMessage(mw.Msg.GetTopic(), mw.Msg.ID)
		g.routingFn(mw.Msg)
	}, gossip.DefaultMsgIDFn)

	valRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[pubsub.ValidationResult]) {
		// hitMap.addValidation(mw.Msg.GetTopic())
		res := g.validationFn(mw.Peer, mw.Msg)
		mw.Result = res
	}, gossip.DefaultMsgIDFn)

	return cfg, msgRouter, valRouter, name
}
