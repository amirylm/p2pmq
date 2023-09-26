package grpcapi

import (
	"context"
	"io"
	"sync"

	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidationRouterImpl is an implementation of ValidationRouterServer.
type ValidationRouterImpl struct {
	proto.ValidationRouterServer

	pubsub core.Pubsuber
	q      chan *proto.Message

	results validationResults
}

// NewValidationRouterServer creates a new ValidationRouterServer instance.
func NewValidationRouterServer(ctrl core.Pubsuber, qSize int) *ValidationRouterImpl {
	return &ValidationRouterImpl{
		pubsub:  ctrl,
		q:       make(chan *proto.Message, qSize),
		results: *newValidationResults(),
	}
}

// TODO: handle duplicated msg id
func (r *ValidationRouterImpl) PushWait(ctx context.Context, next *core.MsgWrapper[pubsub.ValidationResult]) pubsub.ValidationResult {
	mid := next.Msg.ID
	wait := r.results.add(mid)
	defer r.results.cancel(mid)

	select {
	case r.q <- &proto.Message{
		MessageId: mid,
		Topic:     next.Msg.GetTopic(),
		Data:      next.Msg.GetData(),
	}:
	default:
		return pubsub.ValidationIgnore
	}

	select {
	case res := <-wait:
		return res
	case <-ctx.Done():
		return pubsub.ValidationIgnore
	}
}

// Handle implements the Handle RPC method.
func (s *ValidationRouterImpl) Handle(stream proto.ValidationRouter_HandleServer) error {
	go s.receiver(stream)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case next := <-s.q:
			if next == nil {
				return nil
			}
			err := stream.Send(next)
			if err != nil {
				return streamErr(err)
			}
		}
	}
}

func (s *ValidationRouterImpl) receiver(stream proto.ValidationRouter_HandleServer) {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := stream.Recv()
			if err != nil {
				if err == io.EOF || err == context.Canceled { // stream closed
					return
				}
				continue
			}
			if res == nil || res.GetMsg() == nil {
				// TODO: handle? backoff?
				continue
			}
			s.results.resolve(res.GetMsg().GetMessageId(), convertResult(res.GetResult()))
		}
	}
}

type validationResults struct {
	resultsLock sync.RWMutex
	results     map[string]chan pubsub.ValidationResult
}

func newValidationResults() *validationResults {
	return &validationResults{
		results: make(map[string]chan pubsub.ValidationResult),
	}
}

func (vr *validationResults) cancel(mid string) {
	vr.resultsLock.Lock()
	vr.resultsLock.Unlock()

	_, ok := vr.results[mid]
	if ok {
		vr.results[mid] = nil
	}
}

func (vr *validationResults) resolve(mid string, res pubsub.ValidationResult) {
	vr.resultsLock.RLock()
	c, ok := vr.results[mid]
	vr.resultsLock.RUnlock()
	if ok {
		select {
		case c <- res:
		default:
		}
	}
}

func (vr *validationResults) add(mid string) chan pubsub.ValidationResult {
	vr.resultsLock.Lock()
	defer vr.resultsLock.Unlock()

	c := make(chan pubsub.ValidationResult)
	vr.results[mid] = c

	return c
}

func streamErr(err error) error {
	if err == io.EOF || err == context.Canceled { // stream closed
		return nil
	}
	return status.Error(codes.Internal, err.Error())
}

func convertResult(res proto.ValidationResult) pubsub.ValidationResult {
	var vr pubsub.ValidationResult
	switch res {
	case proto.ValidationResult_ACCEPT:
		vr = pubsub.ValidationAccept
	case proto.ValidationResult_REJECT:
		vr = pubsub.ValidationReject
	default:
		vr = pubsub.ValidationIgnore
	}
	return vr
}
