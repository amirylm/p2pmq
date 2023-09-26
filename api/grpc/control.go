package grpcapi

import (
	"context"

	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ControlServiceImpl is an implementation of ControlServiceServer.
type ControlServiceImpl struct {
	proto.ControlServiceServer

	pubsub core.Pubsuber
}

// NewControlServiceServer creates a new ControlServiceServer instance.
func NewControlServiceServer(ps core.Pubsuber) *ControlServiceImpl {
	return &ControlServiceImpl{pubsub: ps}
}

// Publish implements the Publish RPC method.
func (s *ControlServiceImpl) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	if err := s.pubsub.Publish(ctx, req.GetTopic(), req.GetData()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.PublishResponse{}, nil
}

// Subscribe implements the Subscribe RPC method.
func (s *ControlServiceImpl) Subscribe(ctx context.Context, req *proto.SubscribeRequest) (*proto.SubscribeResponse, error) {
	if err := s.pubsub.Subscribe(ctx, req.GetTopic()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.SubscribeResponse{}, nil
}

// Unsubscribe implements the Unsubscribe RPC method.
func (s *ControlServiceImpl) Unsubscribe(_ context.Context, req *proto.SubscribeRequest) (*proto.SubscribeResponse, error) {
	if err := s.pubsub.Unsubscribe(req.GetTopic()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.SubscribeResponse{}, nil
}
