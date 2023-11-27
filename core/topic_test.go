package core

import (
	"testing"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/require"
)

func TestTopicManager(t *testing.T) {
	tm := newTopicManager()

	t.Run("non existing", func(t *testing.T) {
		topicName := "dummy"
		require.False(t, tm.upgradeTopic(topicName, &pubsub.Topic{}))
		require.False(t, tm.addSub(topicName, &pubsub.Subscription{}))
		require.Nil(t, tm.getTopicWrapper(topicName))
		require.Nil(t, tm.getSub(topicName))
	})

	t.Run("happy path", func(t *testing.T) {
		topicName := "dummy"
		tm.joiningTopic(topicName)
		require.True(t, tm.upgradeTopic(topicName, &pubsub.Topic{}))
		require.NotNil(t, tm.getTopicWrapper(topicName))
		require.True(t, tm.addSub(topicName, &pubsub.Subscription{}))
		require.NotNil(t, tm.getSub(topicName))
	})
}
