package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amirylm/p2pmq/commons"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"go.uber.org/zap"
)

func (d *Daemon) setupPubsubRouter(ctx context.Context, cfg commons.Config) error {
	opts := []pubsub.Option{
		pubsub.WithMessageSigning(false),
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign),
		pubsub.WithGossipSubParams(gossipSubParams(cfg.Pubsub.Overlay)),
		pubsub.WithMessageIdFn(msgIDSha256(20)),
	}

	if cfg.Pubsub.MaxMessageSize > 0 {
		opts = append(opts, pubsub.WithMaxMessageSize(cfg.Pubsub.MaxMessageSize))
	}

	if overlay := cfg.Pubsub.Overlay; overlay != nil && overlay.SeenTtl.Milliseconds() > 0 {
		opts = append(opts, pubsub.WithSeenMessagesTTL(cfg.Pubsub.Overlay.SeenTtl))
	}

	denylist := pubsub.NewMapBlacklist()
	opts = append(opts, pubsub.WithBlacklist(denylist))

	// pubsub.WithDefaultValidator() // TODO: check

	if cfg.Pubsub.SubFilter != nil {
		re, err := regexp.Compile(cfg.Pubsub.SubFilter.Pattern)
		if err != nil {
			return err
		}
		sf := pubsub.NewRegexpSubscriptionFilter(re)
		if cfg.Pubsub.SubFilter.Limit > 0 {
			sf = pubsub.WrapLimitSubscriptionFilter(sf, cfg.Pubsub.SubFilter.Limit)
		}
		opts = append(opts, pubsub.WithSubscriptionFilter(sf))
	}

	if cfg.Pubsub.Trace {
		opts = append(opts, pubsub.WithEventTracer(newPubsubTracer(d.lggr.Named("PubsubTracer"))))
	}

	ps, err := pubsub.NewGossipSub(ctx, d.host, opts...)
	if err != nil {
		return err
	}
	d.pubsub = ps
	d.denylist = denylist
	d.manager.topics = make(map[string]*topicWrapper)

	return nil
}

func (d *Daemon) Publish(ctx context.Context, topicName string, data []byte) error {
	topic, err := d.tryJoin(topicName)
	if err != nil {
		return err
	}
	// d.lggr.Debugw("publishing on topic", "topic", topicName, "data", string(data))
	return topic.Publish(ctx, data)
}

func (d *Daemon) Leave(topicName string) error {
	tw := d.manager.getTopicWrapper(topicName)
	state := tw.state.Load()
	switch state {
	case topicStateJoined, topicStateErr:
		err := tw.topic.Close()
		if err != nil {
			tw.state.Store(topicStateErr)
			return err
		}
		tw.state.Store(topicStateUnknown)
	default:
	}
	return nil
}

func (d *Daemon) Unsubscribe(topicName string) {
	tw := d.manager.getTopicWrapper(topicName)
	if tw.state.Load() != topicStateUnknown {
		return
	}
	tw.sub.Cancel()
}

func (d *Daemon) Subscribe(ctx context.Context, topicName string) error {
	topic, err := d.tryJoin(topicName)
	if err != nil {
		return err
	}
	sub, err := d.trySubscribe(topic)
	if err != nil {
		return err
	}
	if sub == nil {
		// already subscribed
		return nil
	}
	go d.listenSubscription(sub)

	return nil
}

func (d *Daemon) listenSubscription(sub *pubsub.Subscription) {
	ctx, cancel := context.WithCancel(d.ctx)
	defer cancel()

	d.lggr.Debugw("listening on topic", "topic", sub.Topic())

	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			if err == pubsub.ErrSubscriptionCancelled || ctx.Err() != nil {
				return
			}
			d.lggr.Warnw("failed to get next msg for subscription", "err", err, "topic", sub.Topic())
			// backoff
			time.Sleep(time.Second)
			continue
		}
		if msg == nil {
			continue
		}
		if err := d.router.Handle(msg); err != nil {
			d.lggr.Warnw("failed to handle next msg for subscription", "err", err, "topic", sub.Topic())
		}
	}
}

func (d *Daemon) tryJoin(topicName string) (*pubsub.Topic, error) {
	topicW := d.manager.getTopicWrapper(topicName)
	if topicW != nil {
		if topicW.state.Load() == topicStateJoining {
			return nil, fmt.Errorf("already tring to join topic %s", topicName)
		}
		return topicW.topic, nil
	}
	d.manager.joiningTopic(topicName)
	topic, err := d.pubsub.Join(topicName)
	if err != nil {
		return nil, err
	}
	d.manager.upgradeTopic(topicName, topic)

	return topic, nil
}

func (d *Daemon) trySubscribe(topic *pubsub.Topic) (sub *pubsub.Subscription, err error) {
	topicName := topic.String()
	sub = d.manager.getSub(topicName)
	if sub != nil {
		return nil, nil
	}
	// TODO: create []pubsub.SubOpt
	sub, err = topic.Subscribe()
	if err != nil {
		return nil, err
	}
	d.manager.addSub(topicName, sub)
	return sub, nil
}

// msgIDSha256 uses sha256 hash of the message content
func msgIDSha256(size int) pubsub.MsgIdFunction {
	return func(pmsg *pubsub_pb.Message) string {
		msg := pmsg.GetData()
		if len(msg) == 0 {
			return ""
		}
		// TODO: optimize, e.g. by using a pool of hashers
		h := sha256.Sum256(msg)
		return hex.EncodeToString(h[:size])
	}
}

func gossipSubParams(overlayp *commons.OverlayParams) pubsub.GossipSubParams {
	gsCfg := pubsub.DefaultGossipSubParams()
	var overlay commons.OverlayParams
	if overlayp != nil {
		overlay = *overlayp
	}
	if overlay.D > 0 {
		gsCfg.D = int(overlay.D)
	}
	if overlay.Dlow > 0 {
		gsCfg.Dlo = int(overlay.Dlow)
	}
	if overlay.Dhi > 0 {
		gsCfg.Dhi = int(overlay.Dhi)
	}
	if overlay.Dlazy > 0 {
		gsCfg.Dlazy = int(overlay.Dlazy)
	}
	if overlay.McacheGossip > 0 {
		gsCfg.MaxIHaveMessages = int(overlay.McacheGossip)
	}
	if overlay.McacheLen > 0 {
		gsCfg.MaxIHaveLength = int(overlay.McacheLen)
	}
	if fanoutTtl := overlay.FanoutTtl; fanoutTtl.Milliseconds() > 0 {
		gsCfg.FanoutTTL = fanoutTtl
	}
	if heartbeat := overlay.HeartbeatInterval; heartbeat.Milliseconds() > 0 {
		gsCfg.HeartbeatInterval = heartbeat
	}
	return gsCfg
}

type pubsubManager struct {
	lock   sync.RWMutex
	topics map[string]*topicWrapper
}

type topicWrapper struct {
	state atomic.Int32
	topic *pubsub.Topic
	sub   *pubsub.Subscription
}

const (
	topicStateUnknown = int32(0)
	topicStateJoining = int32(1)
	topicStateJoined  = int32(2)
	topicStateErr     = int32(10)
)

func (pm *pubsubManager) joiningTopic(name string) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	tw := &topicWrapper{}
	tw.state.Store(topicStateJoining)
	pm.topics[name] = tw
}

func (pm *pubsubManager) upgradeTopic(name string, topic *pubsub.Topic) bool {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	tw, ok := pm.topics[name]
	if !ok {
		return false
	}
	if !tw.state.CompareAndSwap(topicStateJoining, topicStateJoined) {
		return false
	}
	tw.topic = topic
	pm.topics[name] = tw

	return true
}

func (pm *pubsubManager) getTopicWrapper(topic string) *topicWrapper {
	pm.lock.RLock()
	defer pm.lock.RUnlock()

	t, ok := pm.topics[topic]
	if !ok {
		return nil
	}
	return t
}

func (pm *pubsubManager) addSub(name string, sub *pubsub.Subscription) bool {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	tw, ok := pm.topics[name]
	if ok {
		// TODO: enable multiple subscriptions per topic
		return false
	}
	tw.sub = sub

	return true
}

func (pm *pubsubManager) getSub(topic string) *pubsub.Subscription {
	pm.lock.RLock()
	defer pm.lock.RUnlock()

	tw, ok := pm.topics[topic]
	if !ok {
		return nil
	}
	return tw.sub
}

// psTracer helps to trace pubsub events, implements pubsublibp2p.EventTracer
type psTracer struct {
	lggr *zap.SugaredLogger
}

// NewTracer creates an instance of pubsub tracer
func newPubsubTracer(lggr *zap.SugaredLogger) pubsub.EventTracer {
	return &psTracer{
		lggr: lggr.Named("PubsubTracer"),
	}
}

// Trace handles events, implementation of pubsub.EventTracer
func (pst *psTracer) Trace(evt *pubsub_pb.TraceEvent) {
	eType := evt.GetType().String()
	pst.lggr.Debugw("pubsub event", "type", eType)
}
