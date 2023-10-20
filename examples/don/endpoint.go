package don

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcEndPoint string

func (gep GrpcEndPoint) Connect() (*grpc.ClientConn, error) {
	return grpc.Dial(string(gep), grpc.WithTransportCredentials(insecure.NewCredentials()))
}
