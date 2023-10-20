package don

import (
	"context"
	"io"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

type MsgConsumer interface {
	Start(context.Context) error
	Subscribe(ctx context.Context, dons ...string) error
	Stop()
}

type msgConsumer struct {
	threadCtrl utils.ThreadControl
	grpc       GrpcEndPoint
	reports    *ReportBuffer
}

func NewMsgConsumer(reports *ReportBuffer, grpc GrpcEndPoint) MsgConsumer {
	return &msgConsumer{
		threadCtrl: utils.NewThreadControl(),
		grpc:       grpc,
		reports:    reports,
	}
}

func (c *msgConsumer) Subscribe(ctx context.Context, dons ...string) error {
	conn, err := c.grpc.Connect()
	if err != nil {
		return err
	}
	controlClient := proto.NewControlServiceClient(conn)
	for _, did := range dons {
		_, err = controlClient.Subscribe(ctx, &proto.SubscribeRequest{
			Topic: donTopic(did),
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
			r, err := UnmarshalReport(msg.GetData())
			if err != nil || r == nil {
				// bad encoding, not expected
				continue
			}
			if c.reports.Add(r.Src, *r) {
				// TODO: notify new report
				continue
			}
		}
	}
}

func (v *msgConsumer) Stop() {
	v.threadCtrl.Close()
}
