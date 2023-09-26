package grpcapi

import (
	"fmt"
	"net"

	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func NewServices(ps core.Pubsuber, qSize int) (*ControlServiceImpl, *MsgRouterImpl, *ValidationRouterImpl) {
	controlServiceServer := NewControlServiceServer(ps)
	msgRouterServer := NewMsgRouterServer(qSize)
	valRouterServer := NewValidationRouterServer(ps, qSize)

	return controlServiceServer, msgRouterServer, valRouterServer
}

func NewGrpcServer(controlService *ControlServiceImpl, msgRouter *MsgRouterImpl, valRouter *ValidationRouterImpl) *grpc.Server {
	grpcServer := grpc.NewServer()

	proto.RegisterControlServiceServer(grpcServer, controlService)
	proto.RegisterMsgRouterServer(grpcServer, msgRouter)
	proto.RegisterValidationRouterServer(grpcServer, valRouter)

	return grpcServer
}

func ListenGrpc(s *grpc.Server, grpcPort int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		return errors.Wrap(err, "could not create TCP listener")
	}
	defer s.Stop()
	if err := s.Serve(lis); err != nil {
		return errors.Wrap(err, "could not serve grpc")
	}
	return nil
}
