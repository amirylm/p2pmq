package core

import (
	"sync"
	"sync/atomic"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type topicManager struct {
	lock   sync.RWMutex
	topics map[string]*topicWrapper
}

func newTopicManager() *topicManager {
	return &topicManager{
		topics: make(map[string]*topicWrapper),
	}
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

func (tm *topicManager) joiningTopic(name string) {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	tw := &topicWrapper{}
	tw.state.Store(topicStateJoining)
	tm.topics[name] = tw
}

func (tm *topicManager) upgradeTopic(name string, topic *pubsub.Topic) bool {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	tw, ok := tm.topics[name]
	if !ok {
		return false
	}
	if !tw.state.CompareAndSwap(topicStateJoining, topicStateJoined) {
		return false
	}
	tw.topic = topic
	tm.topics[name] = tw

	return true
}

func (tm *topicManager) getTopicWrapper(topic string) *topicWrapper {
	tm.lock.RLock()
	defer tm.lock.RUnlock()

	t, ok := tm.topics[topic]
	if !ok {
		return nil
	}
	return t
}

func (tm *topicManager) addSub(name string, sub *pubsub.Subscription) bool {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	tw, ok := tm.topics[name]
	if !ok {
		return false
	}
	tw.sub = sub

	return true
}

func (tm *topicManager) getSub(topic string) *pubsub.Subscription {
	tm.lock.RLock()
	defer tm.lock.RUnlock()

	tw, ok := tm.topics[topic]
	if !ok {
		return nil
	}
	return tw.sub
}
