package core

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Router[T any] interface {
	// Handle is an async and doesn't wait for response
	Handle(pctx context.Context, p peer.ID, msg *pubsub.Message) error
	// Handle waits for a response, will block the caller
	HandleWait(pctx context.Context, p peer.ID, msg *pubsub.Message) (T, error)
}

type Pubsuber interface {
	Publish(ctx context.Context, topicName string, data []byte) error
	Leave(topicName string) error
	Subscribe(ctx context.Context, topicName string) error
	Unsubscribe(topicName string) error
}
