package client

import (
	"context"
	"io"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

type Consumer interface {
	Start(context.Context) error
	Subscribe(ctx context.Context, topics ...string) error
	Stop()
}

type MsgHandler func(*proto.Message) error

type msgConsumer struct {
	threadCtrl utils.ThreadControl
	grpc       GrpcEndPoint
	handler    MsgHandler
}

func NewConsumer(grpc GrpcEndPoint, handler MsgHandler) Consumer {
	return &msgConsumer{
		threadCtrl: utils.NewThreadControl(),
		grpc:       grpc,
		handler:    handler,
	}
}

func (c *msgConsumer) Subscribe(ctx context.Context, topics ...string) error {
	conn, err := c.grpc.Connect()
	if err != nil {
		return err
	}
	controlClient := proto.NewControlServiceClient(conn)
	for _, topic := range topics {
		_, err = controlClient.Subscribe(ctx, &proto.SubscribeRequest{
			Topic: topic,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *msgConsumer) Start(ctx context.Context) error {
	conn, err := c.grpc.Connect()
	if err != nil {
		return err
	}
	msgRouter := proto.NewMsgRouterClient(conn)
	routerClient, err := msgRouter.Listen(ctx, &proto.ListenRequest{})
	if err != nil {
		return err
	}

	msgQ := make(chan *proto.Message, 128)

	c.threadCtrl.Go(func(ctx context.Context) {
		defer close(msgQ)

		for ctx.Err() == nil {
			msg, err := routerClient.Recv()
			if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
				return
			}
			select {
			case msgQ <- msg:
			default:
				// dropped, TODO: handle or log
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-msgQ:
			if msg == nil {
				return nil
			}
			err := c.handler(msg)
			if err != nil {
				// TODO: handle or log
				continue
			}
		}
	}
}

func (v *msgConsumer) Stop() {
	v.threadCtrl.Close()
}
