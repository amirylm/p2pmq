package core

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core/gossip"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	peerScoreInspectionInterval = time.Minute
)

func (c *Controller) setupPubsubRouter(ctx context.Context, cfg commons.Config) error {
	msgID := gossip.DefaultMsgIDFn
	if cfg.Pubsub.MsgIDFnConfig != nil {
		msgID = gossip.MsgIDFn(gossip.MsgIDFuncType(cfg.Pubsub.MsgIDFnConfig.Type), gossip.MsgIDSize(cfg.Pubsub.MsgIDFnConfig.Size))
	}
	opts := []pubsub.Option{
		pubsub.WithMessageSigning(false),
		pubsub.WithMessageSignaturePolicy(pubsub.StrictNoSign),
		pubsub.WithGossipSubParams(gossip.GossipSubParams(cfg.Pubsub.Overlay)),
		pubsub.WithMessageIdFn(msgID),
	}

	if cfg.Pubsub.Scoring != nil {
		opts = append(opts, pubsub.WithPeerScore(gossip.PeerScores(*cfg.Pubsub)))
		opts = append(opts, pubsub.WithPeerScoreInspect(c.inspectPeerScores, peerScoreInspectionInterval))
	}

	if cfg.Pubsub.MaxMessageSize > 0 {
		opts = append(opts, pubsub.WithMaxMessageSize(cfg.Pubsub.MaxMessageSize))
	}

	if overlay := cfg.Pubsub.Overlay; overlay != nil && overlay.SeenTtl.Milliseconds() > 0 {
		opts = append(opts, pubsub.WithSeenMessagesTTL(cfg.Pubsub.Overlay.SeenTtl))
	}

	denylist := pubsub.NewMapBlacklist()
	opts = append(opts, pubsub.WithBlacklist(denylist))

	if cfg.Pubsub.SubFilter != nil {
		re, err := regexp.Compile(cfg.Pubsub.SubFilter.Pattern)
		if err != nil {
			return err
		}
		sf := pubsub.NewRegexpSubscriptionFilter(re)
		if cfg.Pubsub.SubFilter.Limit > 0 {
			sf = pubsub.WrapLimitSubscriptionFilter(sf, cfg.Pubsub.SubFilter.Limit)
		}
		c.subFilter = sf
		opts = append(opts, pubsub.WithSubscriptionFilter(sf))
	}

	if cfg.Pubsub.Trace != nil {
		var jtracer pubsub.EventTracer
		if len(cfg.Pubsub.Trace.JsonFile) > 0 {
			var err error
			jtracer, err = pubsub.NewJSONTracer(cfg.Pubsub.Trace.JsonFile)
			if err != nil {
				return err
			}
		}
		tracer := newPubsubTracer(c.lggr.Named("PubsubTracer"), cfg.Pubsub.Trace.Debug, cfg.Pubsub.Trace.Skiplist, jtracer)
		c.psTracer = tracer.(*psTracer)
		opts = append(opts, pubsub.WithEventTracer(tracer))
		// TODO: config?
		opts = append(opts, pubsub.WithAppSpecificRpcInspector(func(p peer.ID, rpc *pubsub.RPC) error {
			c.pubsubRpcCounter.Add(1)
			return nil
		}))
	}

	ps, err := pubsub.NewGossipSub(ctx, c.host, opts...)
	if err != nil {
		return err
	}
	c.pubsub = ps
	c.denylist = denylist
	c.topicManager.topics = make(map[string]*topicWrapper)

	return nil
}

func (c *Controller) Publish(ctx context.Context, topicName string, data []byte) error {
	topic, err := c.tryJoin(topicName)
	if err != nil {
		return err
	}
	c.lggr.Debugw("publishing on topic", "topic", topicName, "data", string(data))
	return topic.Publish(ctx, data)
}

func (c *Controller) Leave(topicName string) error {
	tw := c.topicManager.getTopicWrapper(topicName)
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

func (c *Controller) Unsubscribe(topicName string) error {
	tw := c.topicManager.getTopicWrapper(topicName)
	if tw.state.Load() == topicStateUnknown {
		return nil // TODO: topic not found?
	}
	tw.sub.Cancel()
	return nil
}

func (c *Controller) Subscribe(ctx context.Context, topicName string) error {
	topic, err := c.tryJoin(topicName)
	if err != nil {
		return err
	}
	sub, err := c.trySubscribe(topic)
	if err != nil {
		return err
	}
	if sub == nil {
		// already subscribed
		return nil
	}
	c.threadControl.Go(func(ctx context.Context) {
		c.listenSubscription(ctx, sub)
	})

	return nil
}

func (c *Controller) Relay(topicName string) error {
	topic, err := c.tryJoin(topicName)
	if err != nil {
		return err
	}
	cancel, err := topic.Relay()
	if err != nil {
		return err
	}
	c.topicManager.setTopicRelayCancelFn(topicName, cancel)
	return nil
}

func (c *Controller) Unrelay(topicName string) error {
	tw := c.topicManager.getTopicWrapper(topicName)
	if tw.state.Load() == topicStateUnknown {
		return nil // TODO: topic not found?
	}
	tw.relayCancelFn()
	return nil
}

func (c *Controller) listenSubscription(ctx context.Context, sub *pubsub.Subscription) {
	c.lggr.Debugw("listening on topic", "topic", sub.Topic())

	for ctx.Err() == nil {
		msg, err := sub.Next(ctx)
		if err != nil {
			if err == pubsub.ErrSubscriptionCancelled || ctx.Err() != nil {
				return
			}
			c.lggr.Warnw("failed to get next msg for subscription", "err", err, "topic", sub.Topic())
			// backoff
			time.Sleep(time.Second) // TODO: jitter
			continue
		}
		if msg == nil {
			continue
		}
		// if msg.ReceivedFrom == c.host.ID() {
		// 	continue
		// }
		if err := c.msgRouter.Handle(ctx, msg.ReceivedFrom, msg); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.lggr.Warnw("failed to handle next msg for subscription", "err", err, "topic", sub.Topic())
		}
	}
}

func (c *Controller) tryJoin(topicName string) (*pubsub.Topic, error) {
	topicW := c.topicManager.getTopicWrapper(topicName)
	if topicW != nil {
		if topicW.state.Load() == topicStateJoining {
			return nil, fmt.Errorf("already tring to join topic %s", topicName)
		}
		return topicW.topic, nil
	}
	c.topicManager.joiningTopic(topicName)
	opts := []pubsub.TopicOpt{}
	cfg, ok := c.cfg.Pubsub.GetTopicConfig(topicName)
	if ok {
		if cfg.MsgIDFnConfig != nil {
			msgID := gossip.MsgIDFn(gossip.MsgIDFuncType(cfg.MsgIDFnConfig.Type), gossip.MsgIDSize(cfg.MsgIDFnConfig.Size))
			opts = append(opts, pubsub.WithTopicMessageIdFn(msgID))
		}
	}
	topic, err := c.pubsub.Join(topicName, opts...)
	if err != nil {
		return nil, err
	}
	c.topicManager.upgradeTopic(topicName, topic)

	if cfg.MsgValidator != nil || c.cfg.Pubsub.MsgValidator != nil {
		msgValConfig := (&commons.MsgValidationConfig{}).Defaults(c.cfg.Pubsub.MsgValidator)
		if cfg.MsgValidator != nil { // specific topic validator config
			msgValConfig = msgValConfig.Defaults(cfg.MsgValidator)
		}
		valOpts := []pubsub.ValidatorOpt{
			pubsub.WithValidatorInline(false),
			pubsub.WithValidatorTimeout(msgValConfig.Timeout),
			pubsub.WithValidatorConcurrency(msgValConfig.Concurrency),
		}
		if err := c.pubsub.RegisterTopicValidator(topicName, c.validateMsg, valOpts...); err != nil {
			return topic, err
		}
	}

	return topic, nil
}

func (c *Controller) trySubscribe(topic *pubsub.Topic) (sub *pubsub.Subscription, err error) {
	topicName := topic.String()
	sub = c.topicManager.getSub(topicName)
	if sub != nil {
		return nil, nil
	}
	var opts []pubsub.SubOpt
	cfg, ok := c.cfg.Pubsub.GetTopicConfig(topicName)
	if ok {
		if cfg.BufferSize > 0 {
			opts = append(opts, pubsub.WithBufferSize(cfg.BufferSize))
		}
	}
	sub, err = topic.Subscribe(opts...)
	if err != nil {
		return nil, err
	}
	c.topicManager.addSub(topicName, sub)
	return sub, nil
}

func (c *Controller) validateMsg(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
	if p.String() == c.host.ID().String() {
		// accept own messages
		return pubsub.ValidationAccept
	}
	res, err := c.valRouter.HandleWait(ctx, p, msg)
	if err != nil {
		if ctx.Err() == nil {
			c.lggr.Warnw("failed to handle msg", "err", err)
		}
		return pubsub.ValidationIgnore
	}
	return res
}

func (c *Controller) inspectPeerScores(map[peer.ID]*pubsub.PeerScoreSnapshot) {
	// TODO
}
