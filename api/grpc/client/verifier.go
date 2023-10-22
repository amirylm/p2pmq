package client

import (
	"context"
	"io"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

type MsgValidator func(*proto.Message) proto.ValidationResult

type Verifier interface {
	Start(context.Context) error
	Stop()

	Process(*proto.Message) proto.ValidationResult
}

type verifier struct {
	threadCtrl utils.ThreadControl
	grpc       GrpcEndPoint
	validator  MsgValidator
}

func NewVerifier(grpc GrpcEndPoint, validator MsgValidator) Verifier {
	return &verifier{
		threadCtrl: utils.NewThreadControl(),
		grpc:       grpc,
		validator:  validator,
	}
}

func (v *verifier) Start(ctx context.Context) error {
	conn, err := v.grpc.Connect()
	if err != nil {
		return err
	}
	valRouter := proto.NewValidationRouterClient(conn)
	routerClient, err := valRouter.Handle(ctx)
	if err != nil {
		return err
	}

	valQ := make(chan *proto.Message, 1)

	v.threadCtrl.Go(func(ctx context.Context) {
		defer close(valQ)

		for ctx.Err() == nil {
			msg, err := routerClient.Recv()
			if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
				return
			}
			select {
			case <-ctx.Done():
				return
			case valQ <- msg:
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		case next := <-valQ:
			if next == nil {
				return nil
			}
			res := &proto.ValidatedMessage{
				Result: v.Process(next),
				Msg:    next,
			}
			routerClient.Send(res)
		}
	}
}

func (v *verifier) Stop() {
	v.threadCtrl.Close()
}

func (v *verifier) Process(msg *proto.Message) proto.ValidationResult {
	return v.validator(msg)
}
