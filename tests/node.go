package tests

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/amirylm/p2pmq/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type nodeClient struct {
	grpcAddr    string
	defaultConn *grpc.ClientConn

	// opid    string
	reports reportsBuffer

	listenState   atomic.Int32
	validateState atomic.Int32
	connectLock   sync.Mutex
}

func newNodeClient(grpcAddr string) *nodeClient {
	return &nodeClient{
		grpcAddr: grpcAddr,
		// opid:     operatorId,
		reports: reportsBuffer{
			bufferSize: 128,
		},
		listenState:   atomic.Int32{},
		validateState: atomic.Int32{},
	}
}

func (n *nodeClient) subscribe(ctx context.Context, dons ...string) error {
	conn, err := n.connect(true)
	if err != nil {
		return err
	}
	controlClient := proto.NewControlServiceClient(conn)
	for _, did := range dons {
		_, err = controlClient.Subscribe(ctx, &proto.SubscribeRequest{
			Topic: fmt.Sprintf("don.%s", did),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *nodeClient) connect(reuse bool) (*grpc.ClientConn, error) {
	if !reuse {
		return grpc.Dial(c.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	c.connectLock.Lock()
	defer c.connectLock.Unlock()

	if c.defaultConn != nil {
		if c.defaultConn.GetState() == connectivity.Ready {
			return c.defaultConn, nil
		}
		_ = c.defaultConn.Close()
	}
	conn, err := grpc.Dial(c.grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c.defaultConn = conn
	return conn, nil
}

func (c *nodeClient) listen(ctx context.Context) error {
	if !c.listenState.CompareAndSwap(0, 1) {
		return nil
	}
	defer c.listenState.Store(0)

	conn, err := c.connect(false)
	if err != nil {
		return err
	}
	msgRouter := proto.NewMsgRouterClient(conn)
	routerClient, err := msgRouter.Listen(ctx, &proto.ListenRequest{})
	if err != nil {
		return err
	}

	for ctx.Err() == nil {
		msg, err := routerClient.Recv()
		if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
			return nil
		}
		r, err := UnmarshalReport(msg.GetData())
		if err != nil || r == nil {
			// bad encoding, not expected
			continue
		}
		if c.reports.addReport(r.Src, *r) {
			// TODO: notify
			continue
		}
	}
	return nil
}

func (c *nodeClient) validate(ctx context.Context) error {
	if !c.validateState.CompareAndSwap(0, 1) {
		return nil
	}
	defer c.validateState.Store(0)

	conn, err := c.connect(false)
	if err != nil {
		return err
	}
	valRouter := proto.NewValidationRouterClient(conn)
	routerClient, err := valRouter.Handle(ctx)
	if err != nil {
		return err
	}

	valQ := make(chan *proto.Message, 1)

	go func() {
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
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case next := <-valQ:
			r, err := UnmarshalReport(next.GetData())
			if err != nil || r == nil {
				// bad encoding, not expected
				continue
			}
			res := &proto.ValidatedMessage{
				Result: proto.ValidationResult_ACCEPT,
				Msg:    next,
			}
			if r.verify() != nil {
				res.Result = proto.ValidationResult_REJECT
			} else if c.reports.getReport(r.Src, r.SeqNumber) != nil {
				res.Result = proto.ValidationResult_IGNORE
			}
			routerClient.Send(res)
		}
	}
}
