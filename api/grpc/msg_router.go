package grpcapi

import (
	"fmt"

	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/proto"
)

// MsgRouterImpl is an implementation of MsgRouterServer.
type MsgRouterImpl struct {
	proto.MsgRouterServer

	q chan *proto.Message
}

// NewMsgRouterServer creates a new MsgRouterServer instance.
func NewMsgRouterServer(qSize int) *MsgRouterImpl {
	return &MsgRouterImpl{q: make(chan *proto.Message, qSize)}
}

func (r *MsgRouterImpl) Push(next *core.MsgWrapper[error]) error {
	select {
	case r.q <- &proto.Message{
		MessageId: next.Msg.ID,
		Topic:     next.Msg.GetTopic(),
		Data:      next.Msg.GetData(),
	}:
	default:
		return fmt.Errorf("queue is full")
	}
	return nil
}

// Listen implements the Listen RPC method.
func (r *MsgRouterImpl) Listen(req *proto.ListenRequest, stream proto.MsgRouter_ListenServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case next := <-r.q:
			if next == nil {
				return nil
			}
			if err := stream.Send(next); err != nil {
				return streamErr(err)
			}
		}
	}
}
