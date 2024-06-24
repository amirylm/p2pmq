package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
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

var (
	traceEmptySkipList    = []string{}
	traceMsgEventSkipList = []string{
		"ADD_PEER",
		"REMOVE_PEER",
		"JOIN",
		"LEAVE",
		"GRAFT",
		"PRUNE",
		"DROP_RPC",
	}
	traceGossipEventSkipList = []string{
		"ADD_PEER",
		"REMOVE_PEER",
		"JOIN",
		"LEAVE",
	}
)

func TestGossipMsgThroughput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, logging.SetLogLevelRegex("p2pmq", "error"))

	groupsCfgSimple, groupsCfgMedium := testGroupSimple(), testGroupMedium()

	tests := []struct {
		name          string
		n             int
		pubsubConfig  *commons.PubsubConfig
		gen           *testGen
		groupsCfg     map[string]groupCfg
		conns         connectivity
		flows         []flow
		topicsToCheck []string
	}{
		{
			name: "simple_11",
			n:    11,
			gen: &testGen{
				hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace: &commons.PubsubTraceConfig{
						// JsonFile: fmt.Sprintf("../.output/trace/node-%d.json", i+1),
						Skiplist: traceEmptySkipList,
					},
					Overlay: &commons.OverlayParams{
						D:     3,
						Dlow:  2,
						Dhi:   5,
						Dlazy: 3,
					},
				},
			},
			groupsCfg: groupsCfgSimple,
			conns: map[int][]int{
				0: append(append(groupsCfgSimple["a"].ids, groupsCfgSimple["relayers"].ids...),
					5, 7, 9, // group b
				),
				1: append(append(groupsCfgSimple["a"].ids, groupsCfgSimple["relayers"].ids...),
					4, 6, 8, // group b
				),
				2: append(append(groupsCfgSimple["a"].ids, groupsCfgSimple["relayers"].ids...),
					4, 5, 7, // group b
				),
				3: append(append(groupsCfgSimple["a"].ids, groupsCfgSimple["relayers"].ids...),
					4, 6, 8, // group b
				),
				4: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
				5: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
				6: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
				7: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
				8: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
				9: append(groupsCfgSimple["b"].ids, groupsCfgSimple["relayers"].ids...),
			},
			flows: []flow{
				{
					name: "actions a->b",
					events: []flowEvent{
						{
							srcGroup: "a",
							topic:    "b.action.req",
							pattern:  "dummy-request-{{.i}}",
							interval: time.Millisecond * 10,
							wait:     true,
						},
						{
							srcGroup: "b",
							topic:    "b.action.res",
							pattern:  "dummy-response-{{.i}}",
						},
					},
					iterations: 1,
					interval:   time.Millisecond * 250,
				},
				{
					name: "triggers b",
					events: []flowEvent{
						{
							srcGroup: "b",
							topic:    "b.trigger",
							pattern:  "dummy-trigger-{{.group}}-{{.i}}",
							interval: time.Millisecond * 10,
						},
					},
					iterations: 2,
					interval:   time.Millisecond * 10,
				},
			},
		},
		{
			name: "medium_29",
			n:    29,
			gen: &testGen{
				hitMaps:   map[string]*nodeHitMap{},
				routingFn: func(m *pubsub.Message) {},
				validationFn: func(p peer.ID, m *pubsub.Message) pubsub.ValidationResult {
					return pubsub.ValidationAccept
				},
				pubsubConfig: &commons.PubsubConfig{
					MsgValidator: &commons.MsgValidationConfig{},
					Trace: &commons.PubsubTraceConfig{
						// JsonFile: fmt.Sprintf("../.output/trace/node-%d.json", i+1),
						Skiplist: traceEmptySkipList,
					},
					Overlay: &commons.OverlayParams{
						D:     3,
						Dlow:  2,
						Dhi:   5,
						Dlazy: 3,
					},
				},
			},
			groupsCfg: groupsCfgMedium,
			conns: map[int][]int{
				0: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					10, 11, 12, // group b
					20, 21, 22, // group c
				),
				1: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					14, 11, 15, // group b
					24, 21, 25, // group c
				),
				2: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					13, 12, 16, // group b
					23, 22, 26, // group c
				),
				3: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					14, 15, 17, // group b
					24, 25, 27, // group c
				),
				4: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					10, 11, 12, // group b
					20, 21, 22, // group c
				),
				5: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					14, 11, 15, // group b
					21, 20, // group c
				),
				6: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					13, 12, 16, // group b
					23, 22, // group c
				),
				7: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					14, 15, 17, // group b
					25, // group c
				),
				8: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					10, 11, 12, // group b
					20, 21, // group c
				),
				9: append(append(groupsCfgMedium["a"].ids, groupsCfgMedium["relayers"].ids...),
					14, 19, 15, // group b
					24, // group c
				),
				10: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					24, 22, // group c
				),
				11: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					22, // group c
				),
				12: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					22, 23, // group c
				),
				13: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					24, 21, // group c
				),
				14: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					20, // group c
				),
				15: append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
				16: append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
				17: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					22, // group c
				),
				18: append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
				19: append(append(groupsCfgMedium["b"].ids, groupsCfgMedium["relayers"].ids...),
					20, // group c
				),
				20: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				21: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				22: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				23: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				24: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				25: append(groupsCfgMedium["c"].ids, groupsCfgMedium["relayers"].ids...),
				26: groupsCfgMedium["relayers"].ids,
				27: groupsCfgMedium["relayers"].ids,
				28: groupsCfgMedium["relayers"].ids,
			},
			flows: []flow{
				{
					name: "actions a->b",
					events: []flowEvent{
						{
							srcGroup: "a",
							topic:    "b.action.req",
							pattern:  "dummy-request-{{.i}}",
							interval: time.Millisecond * 10,
							wait:     true,
						},
						{
							srcGroup: "b",
							topic:    "b.action.res",
							pattern:  "dummy-response-{{.i}}",
						},
					},
					iterations: 5,
					interval:   time.Millisecond * 250,
				},
				{
					name: "triggers b",
					events: []flowEvent{
						{
							srcGroup: "b",
							topic:    "b.trigger",
							pattern:  "dummy-trigger-{{.group}}-{{.i}}",
							interval: time.Millisecond * 10,
						},
					},
					iterations: 10,
					interval:   time.Millisecond * 10,
				},
				{
					name: "triggers b+c",
					events: []flowEvent{
						{
							srcGroup: "b",
							topic:    "b.trigger",
							pattern:  "xdummy-trigger-{{.group}}-{{.i}}",
							interval: time.Millisecond * 10,
						},
						{
							srcGroup: "c",
							topic:    "c.trigger",
							pattern:  "xdummy-trigger-{{.group}}-{{.i}}",
							interval: time.Millisecond * 10,
						},
					},
					iterations: 10,
					interval:   time.Millisecond * 10,
				},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			ctrls, _, _, done, err := StartControllers(ctx, tc.n, tc.gen)
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

			<-time.After(time.Second * 4) // TODO: avoid timeout

			// starting fresh trace after subscriptions
			startupFaucets := traceFaucets{}
			for _, ctrl := range ctrls {
				startupFaucets.add(ctrl.psTracer.faucets)
				ctrl.psTracer.Reset()
			}
			t.Logf("\n [%s] all trace faucets (startup): %+v\n", tc.name, startupFaucets)

			t.Log("starting flows...")
			for _, f := range tc.flows {
				flow := f
				flowTestName := fmt.Sprintf("%s-x%d", flow.name, flow.iterations)
				t.Run(flowTestName, func(t *testing.T) {
					threadCtrl := utils.NewThreadControl()
					defer threadCtrl.Close()

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
								wg.Add(1)
								threadCtrl.Go(func(ctx context.Context) {
									defer wg.Done()
									args := map[string]interface{}{
										"i":     _i,
										"group": event.srcGroup,
										"ctrl":  ctrl.lggr.Desugar().Name(),
										"flow":  flow.name,
									}
									require.NoError(t, ctrl.Publish(ctx, event.topic, []byte(event.Msg(args))))
								})
							}
							if event.wait {
								wg.Wait()
							}
						}
						<-time.After(flow.interval)
					}

					faucets := traceFaucets{}
					for node, hitMap := range tc.gen.hitMaps {
						for _, topic := range tc.topicsToCheck {
							t.Logf("[%s] %s messages: %d; validations: %d", node, topic, hitMap.messages(topic), hitMap.validations(topic))
						}

						nodeIndex, err := strconv.Atoi(node[5:])
						require.NoError(t, err)
						tracer := ctrls[nodeIndex-1].psTracer
						nodeFaucets := tracer.faucets
						t.Logf("[%s] trace faucets: %+v", node, nodeFaucets)
						faucets.add(nodeFaucets)
						// traceEvents := tracer.Events()
						// eventsJson, err := MarshalTraceEvents(traceEvents)
						// require.NoError(t, err)
						// require.NoError(t, os.WriteFile(fmt.Sprintf("../.output/trace/%s-%s-%s.json", tc.name, node, flow.name), eventsJson, 0644))
						tracer.Reset()
					}
					for _, ctrl := range groups["relayers"] {
						nodeFaucets := ctrl.psTracer.faucets
						t.Logf("[%s] trace faucets: %+v", ctrl.lggr.Desugar().Name(), nodeFaucets)
						faucets.add(nodeFaucets)
					}
					t.Logf("\n [%s/%s] all trace faucets: %+v\n", tc.name, flowTestName, faucets)
				})
			}
		})
	}
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

type groupCfg struct {
	ids    []int
	subs   []string
	relays []string
}

func testGroupSimple() map[string]groupCfg {
	return map[string]groupCfg{
		"a": {
			ids:  []int{0, 1, 2, 3},
			subs: []string{"b.action.res", "b.trigger"},
		},
		"b": {
			ids:  []int{4, 5, 6, 7, 8, 9},
			subs: []string{"b.action.req"},
		},
		"relayers": {
			ids:    []int{10},
			relays: []string{"b.action.req", "b.action.res", "b.trigger"},
		},
	}
}

func testGroupMedium() map[string]groupCfg {
	return map[string]groupCfg{
		"a": {
			ids:  []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			subs: []string{"b.action.res", "b.trigger", "c.trigger"},
		},
		"b": {
			ids:  []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			subs: []string{"b.action.req", "c.trigger"},
		},
		"c": {
			ids:  []int{20, 21, 22, 23, 24, 25},
			subs: []string{"b.trigger"},
		},
		"relayers": {
			ids:    []int{26, 27, 28},
			relays: []string{"b.action.req", "b.action.res", "b.trigger", "c.trigger"},
		},
	}
}

type flowEvent struct {
	srcGroup string
	topic    string
	pattern  string
	interval time.Duration
	wait     bool
}

func (fe flowEvent) Msg(args map[string]interface{}) string {
	tmpl, err := template.New("msg").Parse(fe.pattern)
	if err != nil {
		return ""
	}
	// create io.Buffer for the message and executing template
	sb := new(strings.Builder)
	if err := tmpl.Execute(sb, args); err != nil {
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

type nodeHitMap struct {
	lock      sync.RWMutex
	valHitMap map[string]uint32
	msgHitMap map[string]uint32
}

func (n *nodeHitMap) validations(topic string) uint32 {
	n.lock.RLock()
	defer n.lock.RUnlock()

	return n.valHitMap[topic]
}

func (n *nodeHitMap) messages(topic string) uint32 {
	n.lock.RLock()
	defer n.lock.RUnlock()

	return n.msgHitMap[topic]
}

func (n *nodeHitMap) addValidation(topic string) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.valHitMap[topic] += 1
}

func (n *nodeHitMap) addMessage(topic string) {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.msgHitMap[topic] += 1
}

type testGen struct {
	hitMaps      map[string]*nodeHitMap
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

	hitMap := &nodeHitMap{
		valHitMap: make(map[string]uint32),
		msgHitMap: make(map[string]uint32),
	}
	g.hitMaps[name] = hitMap

	msgRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[error]) {
		hitMap.addMessage(mw.Msg.GetTopic())
		g.routingFn(mw.Msg)
	}, gossip.DefaultMsgIDFn)

	valRouter := NewMsgRouter(1024, 4, func(mw *MsgWrapper[pubsub.ValidationResult]) {
		hitMap.addValidation(mw.Msg.GetTopic())
		res := g.validationFn(mw.Peer, mw.Msg)
		mw.Result = res
	}, gossip.DefaultMsgIDFn)

	return cfg, msgRouter, valRouter, name
}
