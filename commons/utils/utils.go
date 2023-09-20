package utils

import (
	"context"
	"sync"
)

type StartStopOnce struct {
	start sync.Once
	stop  sync.Once
}

func (s *StartStopOnce) StartOnce(fn func()) {
	s.start.Do(fn)
}

func (s *StartStopOnce) StopOnce(fn func()) {
	s.stop.Do(fn)
}

// NOTE: was copied from github.com/smartcontractkit/chainlink
// TODO: consume from github.com/smartcontractkit/chainlink-relay

// A StopChan signals when some work should stop.
type StopChan chan struct{}

// NewCtx returns a background [context.Context] that is cancelled when StopChan is closed.
func (s StopChan) NewCtx() (context.Context, context.CancelFunc) {
	return StopRChan((<-chan struct{})(s)).NewCtx()
}

// Ctx cancels a [context.Context] when StopChan is closed.
func (s StopChan) Ctx(ctx context.Context) (context.Context, context.CancelFunc) {
	return StopRChan((<-chan struct{})(s)).Ctx(ctx)
}

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s StopChan) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	return StopRChan((<-chan struct{})(s)).CtxCancel(ctx, cancel)
}

// A StopRChan signals when some work should stop.
// This version is receive-only.
type StopRChan <-chan struct{}

// NewCtx returns a background [context.Context] that is cancelled when StopChan is closed.
func (s StopRChan) NewCtx() (context.Context, context.CancelFunc) {
	return s.Ctx(context.Background())
}

// Ctx cancels a [context.Context] when StopChan is closed.
func (s StopRChan) Ctx(ctx context.Context) (context.Context, context.CancelFunc) {
	return s.CtxCancel(context.WithCancel(ctx))
}

// CtxCancel cancels a [context.Context] when StopChan is closed.
// Returns ctx and cancel unmodified, for convenience.
func (s StopRChan) CtxCancel(ctx context.Context, cancel context.CancelFunc) (context.Context, context.CancelFunc) {
	go func() {
		select {
		case <-s:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
