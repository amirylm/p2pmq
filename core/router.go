package core

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/commons/utils"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type MsgWrapper[T any] struct {
	Peer      peer.ID
	Msg       *pubsub.Message
	Result    T
	callbackC chan T
}

func (mw *MsgWrapper[T]) callback(res T) {
	if mw.callbackC == nil {
		return
	}
	select {
	case mw.callbackC <- res:
	default:
	}
}

type MsgRouter[T any] interface {
	Handle(pctx context.Context, p peer.ID, msg *pubsub.Message) error
	HandleWait(pctx context.Context, p peer.ID, msg *pubsub.Message) (T, error)

	Start(pctx context.Context)
}

type msgRouter[T any] struct {
	threadControl utils.ThreadControl

	workers chan struct{}
	handler func(*MsgWrapper[T])
	q       chan *MsgWrapper[T]

	msgIDFn pubsub.MsgIdFunction
}

func NewMsgRouter[T any](qSize, workers int, handler func(*MsgWrapper[T]), msgIDFn pubsub.MsgIdFunction) MsgRouter[T] {
	return &msgRouter[T]{
		threadControl: utils.NewThreadControl(),
		workers:       make(chan struct{}, workers),
		handler:       handler,
		q:             make(chan *MsgWrapper[T], qSize),
		msgIDFn:       msgIDFn,
	}
}

func (r *msgRouter[T]) Start(pctx context.Context) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	for {
		select {
		case msg := <-r.q:
			r.dispatchMsg(msg)
		case <-ctx.Done():
			close(r.q)
			return
		}
	}
}

func (r *msgRouter[T]) dispatchMsg(msg *MsgWrapper[T]) {
	r.workers <- struct{}{}
	r.threadControl.Go(func(ctx context.Context) {
		defer func() {
			<-r.workers
		}()
		r.handler(msg)
		msg.callback(msg.Result)
	})
}

func (r *msgRouter[T]) Handle(pctx context.Context, p peer.ID, msg *pubsub.Message) error {
	select {
	case r.q <- &MsgWrapper[T]{
		Peer: p,
		Msg:  msg,
	}:
	default:
		return fmt.Errorf("router queue is full, dropping msg")
	}
	return nil
}

func (r *msgRouter[T]) HandleWait(pctx context.Context, p peer.ID, msg *pubsub.Message) (T, error) {
	var res T
	mw := MsgWrapper[T]{
		Peer:      p,
		Msg:       msg,
		callbackC: make(chan T),
	}
	select {
	case r.q <- &mw:
	default:
		return res, fmt.Errorf("router queue is full, dropping msg")
	}

	select {
	case res = <-mw.callbackC:
	case <-pctx.Done():
		return res, fmt.Errorf("timeout waiting for msg %s", r.msgIDFn(msg.Message))
	}

	return res, nil
}
